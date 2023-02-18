package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/status-im/go-waku/waku/v2/protocol"
	"github.com/status-im/go-waku/waku/v2/protocol/pb"
	"github.com/status-im/go-waku/waku/v2/protocol/relay"
	"github.com/status-im/go-waku/waku/v2/utils"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/migrate"
	"github.com/xmtp/xmtp-node-go/pkg/logging"
	"github.com/xmtp/xmtp-node-go/pkg/metrics"
	"github.com/xmtp/xmtp-node-go/pkg/migrations/messages"
	"github.com/xmtp/xmtp-node-go/pkg/tracing"
	"go.uber.org/zap"
)

const bufferSize = 1024
const maxPageSize = 100
const maxPeersToResume = 5

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

type XmtpStore struct {
	ctx       context.Context
	cancel    func()
	MsgC      chan *protocol.Envelope
	wg        sync.WaitGroup
	db        *sql.DB
	readerDB  *sql.DB
	cleanerDB *sql.DB
	log       *zap.Logger

	started     bool
	statsPeriod time.Duration
	cleaner     CleanerOptions
}

func New(opts ...Option) (*XmtpStore, error) {
	s := new(XmtpStore)
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

func (s *XmtpStore) start() error {
	if s.started {
		return nil
	}
	s.started = true
	s.ctx, s.cancel = context.WithCancel(s.ctx)

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

func (s *XmtpStore) Close() {
	s.started = false

	if s.MsgC != nil {
		close(s.MsgC)
	}

	s.cancel()
	s.wg.Wait()
	s.log.Info("stopped")
}

func (s *XmtpStore) FindMessages(query *pb.HistoryQuery) (res *pb.HistoryResponse, err error) {
	return FindMessages(s.db, query)
}

func (s *XmtpStore) statusMetricsLoop(ctx context.Context) {
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

func (s *XmtpStore) InsertMessage(msg *pb.WakuMessage) (bool, error) {
	env := protocol.NewEnvelope(msg, time.Now().UTC().UnixNano(), relay.DefaultWakuTopic)
	return s.storeMessage(env)
}

func (s *XmtpStore) storeMessage(env *protocol.Envelope) (stored bool, err error) {
	err = tracing.Wrap(s.ctx, "storing message", func(ctx context.Context, span tracing.Span) error {
		tracing.SpanResource(span, "store")
		tracing.SpanType(span, "db")
		err = s.insertMessage(env) // Should the index be stored?
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
		s.log.Debug("message stored",
			zap.String("content_topic", env.Message().ContentTopic),
			zap.Int("size", env.Size()),
			logging.Time("sent", env.Index().SenderTime))
		// This expects me to know the length of the message queue, which I don't now that the store lives in the DB. Setting to 1 for now
		// TODO: re-add this without the waku reference
		// metrics.RecordMessage(s.ctx, "stored", 1)
		tracing.SpanTag(span, "stored", true)
		tracing.SpanTag(span, "content_topic", env.Message().ContentTopic)
		tracing.SpanTag(span, "size", env.Size())
		return nil
	})
	return stored, err
}

func (s *XmtpStore) insertMessage(env *protocol.Envelope) error {
	cursor := env.Index()
	pubsubTopic := env.PubsubTopic()
	message := env.Message()
	shouldExpire := !isXMTP(message)
	sql := "INSERT INTO message (id, receiverTimestamp, senderTimestamp, contentTopic, pubsubTopic, payload, version, should_expire) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)"
	_, err := s.db.Exec(sql, cursor.Digest, cursor.ReceiverTime, message.Timestamp, message.ContentTopic, pubsubTopic, message.Payload, message.Version, shouldExpire)
	return err
}

func (s *XmtpStore) migrate() error {
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

func isXMTP(msg *pb.WakuMessage) bool {
	return strings.HasPrefix(msg.ContentTopic, "/xmtp/")
}

func computeIndex(env *protocol.Envelope) (*pb.Index, error) {
	hash := sha256.Sum256(append([]byte(env.Message().ContentTopic), env.Message().Payload...))
	return &pb.Index{
		Digest:       hash[:],
		ReceiverTime: utils.GetUnixEpoch(),
		SenderTime:   env.Message().Timestamp,
		PubsubTopic:  env.PubsubTopic(),
	}, nil
}
