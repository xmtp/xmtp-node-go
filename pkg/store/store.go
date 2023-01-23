package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/pkg/errors"
	"github.com/status-im/go-waku/waku/v2/protocol"
	"github.com/status-im/go-waku/waku/v2/protocol/pb"
	"github.com/status-im/go-waku/waku/v2/protocol/store"
	"github.com/status-im/go-waku/waku/v2/utils"
	messagev1 "github.com/xmtp/proto/v3/go/message_api/v1"
	"github.com/xmtp/xmtp-node-go/pkg/tracing"
	"go.uber.org/zap"
)

const bufferSize = 1024
const maxPageSize = 100

var (
	ErrMissingLogOption             = errors.New("missing log option")
	ErrMissingHostOption            = errors.New("missing host option")
	ErrMissingDBOption              = errors.New("missing db option")
	ErrMissingMessageProviderOption = errors.New("missing message provider option")
	ErrMissingStatsPeriodOption     = errors.New("missing stats period option")
	ErrMissingCleanerActivePeriod   = errors.New("missing cleaner active period option")
	ErrMissingCleanerPassivePeriod  = errors.New("missing cleaner passive period option")
	ErrMissingCleanerBatchSize      = errors.New("missing cleaner batch size option")
	ErrMissingCleanerRetentionDays  = errors.New("missing cleaner retention days option")
)

type XmtpStore struct {
	ctx         context.Context
	cancel      func()
	MsgC        chan *protocol.Envelope
	wg          sync.WaitGroup
	db          *sql.DB
	readerDB    *sql.DB
	log         *zap.Logger
	host        host.Host
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

	// Required host option.
	if s.host == nil {
		return nil, ErrMissingHostOption
	}
	s.log = s.log.With(zap.String("host", s.host.ID().Pretty()))

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
	}

	s.MsgC = make(chan *protocol.Envelope, bufferSize)

	if s.cleaner.Enable {
		go s.cleanerLoop()
	}

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

	s.log.Info("Store protocol started")
}

func (s *XmtpStore) Stop() {
	s.started = false

	if s.MsgC != nil {
		close(s.MsgC)
	}

	if s.host != nil {
		s.host.RemoveStreamHandler(store.StoreID_v20beta4)
	}

	s.cancel()
	s.wg.Wait()
	s.log.Info("stopped")
}

func (s *XmtpStore) FindMessages(query *pb.HistoryQuery) (res *pb.HistoryResponse, err error) {
	return FindMessages(s.db, query)
}

func (s *XmtpStore) storeMessage(env *messagev1.Envelope) (stored bool, err error) {
	err = tracing.Wrap(s.ctx, "storing message", func(ctx context.Context, span tracing.Span) error {
		tracing.SpanResource(span, "store")
		tracing.SpanType(span, "db")
		// err = s.msgProvider.Put(env) // Should the index be stored?
		// if err != nil {
		// 	tracing.SpanTag(span, "stored", false)
		// 	if err, ok := err.(pgdriver.Error); ok && err.IntegrityViolation() {
		// 		s.log.Debug("storing message", zap.Error(err))
		// 		// metrics.RecordStoreError(s.ctx, "store_duplicate_key")
		// 		return nil
		// 	}
		// 	s.log.Error("storing message", zap.Error(err))
		// 	// metrics.RecordStoreError(s.ctx, "store_failure")
		// 	span.Finish(tracing.WithError(err))
		// 	return err
		// }
		// stored = true
		// s.log.Debug("message stored",
		// 	zap.String("content_topic", env.Message().ContentTopic),
		// 	zap.Int("size", env.Size()),
		// 	zap.Int64("sent", env.Index().SenderTime))
		// // This expects me to know the length of the message queue, which I don't now that the store lives in the DB. Setting to 1 for now
		// // metrics.RecordMessage(s.ctx, "stored", 1)
		// tracing.SpanTag(span, "stored", true)
		// tracing.SpanTag(span, "content_topic", env.Message().ContentTopic)
		// tracing.SpanTag(span, "size", env.Size())
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
