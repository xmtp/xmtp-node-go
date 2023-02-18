package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/status-im/go-waku/waku/v2/protocol"
	"github.com/status-im/go-waku/waku/v2/protocol/pb"
	"github.com/status-im/go-waku/waku/v2/protocol/relay"
	"github.com/status-im/go-waku/waku/v2/protocol/store"
	"github.com/status-im/go-waku/waku/v2/utils"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/xmtp/xmtp-node-go/pkg/logging"
	"github.com/xmtp/xmtp-node-go/pkg/metrics"
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
	ctx         context.Context
	cancel      func()
	MsgC        chan *protocol.Envelope
	wg          sync.WaitGroup
	db          *sql.DB
	readerDB    *sql.DB
	cleanerDB   *sql.DB
	log         *zap.Logger
	msgProvider store.MessageProvider

	started     bool
	statsPeriod time.Duration
	cleaner     CleanerOptions
}

func NewXmtpStore(opts ...Option) (*XmtpStore, error) {
	s := new(XmtpStore)
	for _, opt := range opts {
		opt(s)
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

	// Required db option.
	if s.msgProvider == nil {
		return nil, ErrMissingMessageProviderOption
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

	s.MsgC = make(chan *protocol.Envelope, bufferSize)

	return s, nil
}

func (s *XmtpStore) MessageChannel() chan *protocol.Envelope {
	return s.MsgC
}

func (s *XmtpStore) Start(ctx context.Context) {
	if s.started {
		return
	}
	s.started = true
	s.ctx, s.cancel = context.WithCancel(ctx)

	tracing.GoPanicWrap(s.ctx, &s.wg, "store-status-metrics", func(ctx context.Context) { s.statusMetricsLoop(ctx) })
	s.log.Info("Store protocol started")

	// Start cleaner
	if s.cleaner.Enable {
		go s.cleanerLoop()
	}
}

func (s *XmtpStore) Stop() {
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

func (s *XmtpStore) findLastSeen() (int64, error) {
	res, err := FindMessages(s.db, &pb.HistoryQuery{
		PagingInfo: &pb.PagingInfo{
			Direction: pb.PagingInfo_BACKWARD,
			PageSize:  1,
		},
	})
	if err != nil || len(res.Messages) == 0 {
		return 0, err
	}

	return res.Messages[0].Timestamp, nil
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
		err = s.msgProvider.Put(env) // Should the index be stored?
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

func computeIndex(env *protocol.Envelope) (*pb.Index, error) {
	hash := sha256.Sum256(append([]byte(env.Message().ContentTopic), env.Message().Payload...))
	return &pb.Index{
		Digest:       hash[:],
		ReceiverTime: utils.GetUnixEpoch(),
		SenderTime:   env.Message().Timestamp,
		PubsubTopic:  env.PubsubTopic(),
	}, nil
}

func max(x, y int64) int64 {
	if x > y {
		return x
	}
	return y
}
