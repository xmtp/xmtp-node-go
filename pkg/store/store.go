package store

import (
	"context"
	"strings"
	"sync"
	"time"

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

type Store struct {
	config  *Config
	ctx     context.Context
	cancel  func()
	wg      sync.WaitGroup
	log     *zap.Logger
	started bool
	now     func() time.Time
}

func New(config *Config) (*Store, error) {
	s := &Store{
		config: config,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
	s.ctx, s.cancel = context.WithCancel(context.Background())

	err := config.validate()
	if err != nil {
		return nil, err
	}

	s.log = config.Log.Named("store")

	err = s.start()
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

	err := s.migrate()
	if err != nil {
		return err
	}

	tracing.GoPanicWrap(s.ctx, &s.wg, "store-status-metrics", func(ctx context.Context) { s.metricsLoop(ctx) })
	s.log.Info("Store protocol started")

	// Start cleaner
	if s.config.Cleaner.Enable {
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

func (s *Store) metricsLoop(ctx context.Context) {
	if s.config.MetricsPeriod == 0 {
		s.log.Info("statsPeriod is 0 indicating no metrics loop")
		return
	}
	ticker := time.NewTicker(s.config.MetricsPeriod)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			metrics.EmitStoredMessages(ctx, s.config.ReaderDB, s.log)
		}
	}
}

func (s *Store) InsertMessage(env *messagev1.Envelope) (bool, error) {
	var stored bool
	err := tracing.Wrap(s.ctx, s.log, "storing message", func(ctx context.Context, log *zap.Logger, span tracing.Span) error {
		tracing.SpanResource(span, "store")
		tracing.SpanType(span, "db")
		err := s.insertMessage(env, s.now().UnixNano())
		if err != nil {
			tracing.SpanTag(span, "stored", false)
			if err, ok := err.(pgdriver.Error); ok && err.IntegrityViolation() {
				s.log.Debug("storing message", zap.Error(err))
				return nil
			}
			s.log.Error("storing message", zap.Error(err))
			span.Finish(tracing.WithError(err))
			return err
		}
		stored = true
		s.log.Debug("message stored", zap.String("content_topic", env.ContentTopic), logging.Time("sent", int64(env.TimestampNs)))
		tracing.SpanTag(span, "stored", true)
		tracing.SpanTag(span, "content_topic", env.ContentTopic)
		return nil
	})
	return stored, err
}

func (s *Store) insertMessage(env *messagev1.Envelope, receiverTimestamp int64) error {
	digest := computeDigest(env)
	shouldExpire := !isXMTP(env.ContentTopic)
	sql := "INSERT INTO message (id, receiverTimestamp, senderTimestamp, contentTopic, pubsubTopic, payload, version, should_expire) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)"
	_, err := s.config.DB.Exec(sql, digest, receiverTimestamp, env.TimestampNs, env.ContentTopic, "", env.Message, 0, shouldExpire)
	return err
}

func (s *Store) migrate() error {
	ctx := context.Background()
	db := bun.NewDB(s.config.DB, pgdialect.New())
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
