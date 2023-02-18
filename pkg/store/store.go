package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/migrate"
	messagev1 "github.com/xmtp/proto/v3/go/message_api/v1"
	"github.com/xmtp/xmtp-node-go/pkg/logging"
	"github.com/xmtp/xmtp-node-go/pkg/metrics"
	"github.com/xmtp/xmtp-node-go/pkg/migrations/messages"
	"github.com/xmtp/xmtp-node-go/pkg/tracing"
	"go.uber.org/zap"
)

const maxPageSize = 100

var (
	ErrMissingLogOption             = errors.New("missing log option")
	ErrMissingDBOption              = errors.New("missing db option")
	ErrMissingMessageProviderOption = errors.New("missing message provider option")
	ErrMissingStatsPeriodOption     = errors.New("missing stats period option")
	ErrMissingCleanerActivePeriod   = errors.New("missing cleaner active period option")
	ErrMissingCleanerPassivePeriod  = errors.New("missing cleaner passive period option")
	ErrMissingCleanerBatchSize      = errors.New("missing cleaner batch size option")
	ErrMissingCleanerRetentionDays  = errors.New("missing cleaner retention days option")
	ErrMissingCleanerDBOption       = errors.New("missing cleaner db option")
)

type Store struct {
	ctx       context.Context
	cancel    func()
	wg        sync.WaitGroup
	db        *sql.DB
	readerDB  *sql.DB
	cleanerDB *sql.DB
	log       *zap.Logger

	started     bool
	statsPeriod time.Duration
	cleaner     CleanerOptions
}

func New(opts ...Option) (*Store, error) {
	s := new(Store)
	for _, opt := range opts {
		opt(s)
	}

	if s.ctx == nil {
		s.ctx, s.cancel = context.WithCancel(context.Background())
	}

	// Required logger option.
	if s.log == nil {
		return nil, ErrMissingLogOption
	}
	s.log = s.log.Named("store")

	// Required db option.
	if s.db == nil {
		return nil, ErrMissingDBOption
	}

	// Required reader db option.
	if s.readerDB == nil {
		return nil, ErrMissingDBOption
	}

	// Required cleaner options.
	if s.cleaner.Enable {
		if s.cleaner.ActivePeriod == 0 {
			return nil, ErrMissingCleanerActivePeriod
		}
		if s.cleaner.PassivePeriod == 0 {
			return nil, ErrMissingCleanerPassivePeriod
		}
		if s.cleaner.BatchSize == 0 {
			return nil, ErrMissingCleanerBatchSize
		}
		if s.cleaner.RetentionDays == 0 {
			return nil, ErrMissingCleanerRetentionDays
		}
		if s.cleanerDB == nil {
			return nil, ErrMissingCleanerDBOption
		}
	}

	err := s.start()
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Store) start() error {
	if s.started {
		return nil
	}
	s.started = true
	s.ctx, s.cancel = context.WithCancel(s.ctx)
	s.ctx = logging.With(s.ctx, s.log)

	err := s.migrate()
	if err != nil {
		return err
	}

	tracing.GoPanicWrap(s.ctx, &s.wg, "store-status-metrics", func(ctx context.Context) { s.statusMetricsLoop(ctx) })
	s.log.Info("Store protocol started")

	// Start cleaner
	if s.cleaner.Enable {
		go s.cleanerLoop()
	}

	return nil
}

func (s *Store) Close() {
	s.started = false
	s.cancel()
	s.wg.Wait()
	s.log.Info("stopped")
}

func (s *Store) FindMessages(query *messagev1.QueryRequest) (res *messagev1.QueryResponse, err error) {
	return FindMessages(s.db, query)
}

func (s *Store) statusMetricsLoop(ctx context.Context) {
	if s.statsPeriod == 0 {
		s.log.Info("statsPeriod is 0 indicating no metrics loop")
		return
	}
	ticker := time.NewTicker(s.statsPeriod)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			metrics.EmitStoredMessages(ctx, s.readerDB, s.log)
		}
	}
}

func (s *Store) InsertMessage(env *messagev1.Envelope) (bool, error) {
	return s.storeMessage(env, time.Now().UTC().UnixNano())
}

func (s *Store) storeMessage(env *messagev1.Envelope, ts int64) (stored bool, err error) {
	err = tracing.Wrap(s.ctx, "storing message", func(ctx context.Context, span tracing.Span) error {
		tracing.SpanResource(span, "store")
		tracing.SpanType(span, "db")
		err = s.insertMessage(env, ts) // Should the index be stored?
		if err != nil {
			tracing.SpanTag(span, "stored", false)
			if err, ok := err.(pgdriver.Error); ok && err.IntegrityViolation() {
				s.log.Debug("storing message", zap.Error(err))
				// TODO: re-add this without the waku reference
				// metrics.RecordStoreError(s.ctx, "store_duplicate_key")
				return nil
			}
			s.log.Error("storing message", zap.Error(err))
			// TODO: re-add this without the waku reference
			// metrics.RecordStoreError(s.ctx, "store_failure")
			span.Finish(tracing.WithError(err))
			return err
		}
		stored = true
		s.log.Debug("message stored", zap.String("content_topic", env.ContentTopic), logging.Time("sent", int64(env.TimestampNs)))
		// This expects me to know the length of the message queue, which I don't now that the store lives in the DB. Setting to 1 for now
		// TODO: re-add this without the waku reference
		// metrics.RecordMessage(s.ctx, "stored", 1)
		tracing.SpanTag(span, "stored", true)
		tracing.SpanTag(span, "content_topic", env.ContentTopic)
		return nil
	})
	return stored, err
}

func (s *Store) insertMessage(env *messagev1.Envelope, ts int64) error {
	digest := computeDigest(env)
	shouldExpire := !isXMTP(env.ContentTopic)
	sql := "INSERT INTO message (id, receiverTimestamp, senderTimestamp, contentTopic, pubsubTopic, payload, version, should_expire) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)"
	_, err := s.db.Exec(sql, digest, ts, env.TimestampNs, env.ContentTopic, "", env.Message, 0, shouldExpire)
	return err
}

func (s *Store) migrate() error {
	ctx := context.Background()
	db := bun.NewDB(s.db, pgdialect.New())
	migrator := migrate.NewMigrator(db, messages.Migrations)
	err := migrator.Init(ctx)
	if err != nil {
		return err
	}

	group, err := migrator.Migrate(ctx)
	if err != nil {
		return err
	}
	if group.IsZero() {
		s.log.Info("No new migrations to run")
	}

	return nil
}

func isXMTP(topic string) bool {
	return strings.HasPrefix(topic, "/xmtp/")
}

func computeDigest(env *messagev1.Envelope) []byte {
	digest := sha256.Sum256(append([]byte(env.ContentTopic), env.Message...))
	return digest[:]
}
