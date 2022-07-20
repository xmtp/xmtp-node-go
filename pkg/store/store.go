package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"math"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-msgio/protoio"
	"github.com/pkg/errors"
	"github.com/status-im/go-waku/waku/v2/protocol"
	"github.com/status-im/go-waku/waku/v2/protocol/pb"
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
	ErrMissingHostOption            = errors.New("missing host option")
	ErrMissingDBOption              = errors.New("missing db option")
	ErrMissingMessageProviderOption = errors.New("missing message provider option")
	ErrMissingStatsPeriodOption     = errors.New("missing stats period option")
)

const (
	relayPingContentTopic = "/relay-ping/1/ping/null"
)

type XmtpStore struct {
	ctx         context.Context
	MsgC        chan *protocol.Envelope
	wg          sync.WaitGroup
	db          *sql.DB
	log         *zap.Logger
	host        host.Host
	msgProvider store.MessageProvider

	started         bool
	statsPeriod     time.Duration
	resumePageSize  int
	resumeStartTime int64
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

	// Required db option.
	if s.msgProvider == nil {
		return nil, ErrMissingMessageProviderOption
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
	s.ctx = ctx
	s.host.SetStreamHandler(store.StoreID_v20beta4, s.onRequest)

	tracing.GoPanicWrap(ctx, &s.wg, "store-incoming-messages", func(ctx context.Context) { s.storeIncomingMessages(ctx) })
	tracing.GoPanicWrap(ctx, &s.wg, "store-status-metrics", func(ctx context.Context) { s.statusMetricsLoop(ctx) })
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

	s.wg.Wait()
	s.log.Info("stopped")
}

func (s *XmtpStore) FindMessages(query *pb.HistoryQuery) (res *pb.HistoryResponse, err error) {
	return FindMessages(s.db, query)
}

func (s *XmtpStore) Query(ctx context.Context, query store.Query, opts ...store.HistoryRequestOption) (*store.Result, error) {
	s.log.Named("query").Error("not implemented")
	return nil, errors.New("not implemented")
}

// Next is used to retrieve the next page of rows from a query response.
// If no more records are found, the result will not contain any messages.
// This function is useful for iterating over results without having to manually
// specify the cursor and pagination order and max number of results
func (s *XmtpStore) Next(ctx context.Context, r *store.Result) (*store.Result, error) {
	s.log.Named("next").Error("not implemented")
	return nil, errors.New("not implemented")
}

// Resume retrieves the history of waku messages published on the default waku pubsub topic since the last time the waku store node has been online
// messages are stored in the store node's messages field and in the message db
// the offline time window is measured as the difference between the current time and the timestamp of the most recent persisted waku message
// an offset of 20 second is added to the time window to count for nodes asynchrony
// the history is fetched from one of the peers persisted in the waku store node's peer manager unit
// peerList indicates the list of peers to query from. The history is fetched from the first available peer in this list. Such candidates should be found through a discovery method (to be developed).
// if no peerList is passed, one of the peers in the underlying peer manager unit of the store protocol is picked randomly to fetch the history from. The history gets fetched successfully if the dialed peer has been online during the queried time window.
// the resume proc returns the number of retrieved messages if no error occurs, otherwise returns the error string
func (s *XmtpStore) Resume(ctx context.Context, pubsubTopic string, peers []peer.ID) (int, error) {
	if !s.started {
		return 0, errors.New("can't resume: store has not started")
	}
	log := s.log.With(zap.String("pubsub_topic", pubsubTopic))

	currentTime := utils.GetUnixEpoch()
	offset := int64(20 * time.Second)
	endTime := currentTime + offset

	var startTime int64
	if s.resumeStartTime < 0 {
		lastSeenTime, err := s.findLastSeen()
		if err != nil {
			return 0, errors.Wrap(err, "finding latest timestamp")
		}
		startTime = max(lastSeenTime-offset, 0)
	} else {
		startTime = s.resumeStartTime
	}

	if len(peers) == 0 {
		var err error
		peers, err = selectPeers(s.host, string(store.StoreID_v20beta4), maxPeersToResume, s.log)
		if err != nil {
			return 0, errors.Wrap(err, "selecting peer")
		}
	}
	if len(peers) == 0 {
		return 0, errors.New("no peers")
	}

	log.Info("resuming", zap.Int64("start_time", startTime), zap.Int64("end_time", endTime), zap.Int("page_size", s.resumePageSize))

	req := &pb.HistoryQuery{
		PubsubTopic: pubsubTopic,
		StartTime:   startTime,
		EndTime:     endTime,
		PagingInfo: &pb.PagingInfo{
			PageSize:  uint64(s.resumePageSize),
			Direction: pb.PagingInfo_FORWARD,
		},
	}
	var wg sync.WaitGroup
	var msgCount int
	var msgCountLock sync.RWMutex
	var asyncErr error
	var success bool
	started := time.Now().UTC()
	var latestStoredTimestamp int64
	for _, p := range peers {
		wg.Add(1)
		go func(p peer.ID) {
			defer wg.Done()
			log := s.log.With(logging.HostID("peer", p))

			var msgFnErr error
			// NOTE that we intentionally do not use the ctx passed into the
			// method, since it's created within go-waku with a 20 second
			// timeout. We don't want a timeout on this, and the store-owned
			// context will trigger a cancel on tear down.
			count, err := s.queryPeer(s.ctx, req, p, func(msg *pb.WakuMessage) bool {
				timestamp := utils.GetUnixEpoch()
				stored, err := s.storeMessage(protocol.NewEnvelope(msg, timestamp, req.PubsubTopic))
				if err != nil {
					s.log.Error("storing message", zap.Error(err))
					msgFnErr = err
					return false
				}
				latestStoredTimestamp = timestamp
				return stored
			})
			if err != nil {
				log.Error("querying peer", zap.Error(err))
				asyncErr = err
				return
			}
			if msgFnErr != nil {
				log.Error("querying peer")
				asyncErr = msgFnErr
				return
			}

			success = true

			msgCountLock.Lock()
			defer msgCountLock.Unlock()
			msgCount += count
		}(p)
	}
	wg.Wait()
	log = log.With(zap.Int("count", msgCount), zap.Duration("duration", time.Now().UTC().Sub(started)), zap.Int64("latest_stored_timestamp", latestStoredTimestamp))
	if !success && asyncErr != nil {
		log.Error("resuming", zap.Error(asyncErr))
		return msgCount, asyncErr
	}

	log.Info("resume complete")
	return msgCount, nil
}

func (s *XmtpStore) queryPeer(ctx context.Context, req *pb.HistoryQuery, peerID peer.ID, msgFn func(*pb.WakuMessage) bool) (int, error) {
	c, err := NewClient(
		WithClientLog(s.log),
		WithClientHost(s.host),
		WithClientPeer(peerID),
	)
	if err != nil {
		return 0, err
	}

	msgCount, err := c.Query(ctx, req, func(res *pb.HistoryResponse) (int, bool) {
		var count int
		for _, msg := range res.Messages {
			ok := msgFn(msg)
			if !ok {
				continue
			}
			count++
		}
		return count, true
	})
	if err != nil {
		return 0, err
	}

	return msgCount, nil
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

func (s *XmtpStore) onRequest(stream network.Stream) {
	defer stream.Close()
	_ = tracing.Wrap(s.ctx, "store request", func(ctx context.Context, span tracing.Span) error {
		log := s.log.With(logging.HostID("peer", stream.Conn().RemotePeer()))
		log = tracing.Link(span, log)
		tracing.SpanResource(span, "store")
		tracing.SpanType(span, "p2p")
		tracing.SpanTag(span, "peer", stream.Conn().RemotePeer())

		historyRPCRequest := &pb.HistoryRPC{}
		writer := protoio.NewDelimitedWriter(stream)
		reader := protoio.NewDelimitedReader(stream, math.MaxInt32)

		err := tracing.Wrap(ctx, "reading request", func(ctx context.Context, span tracing.Span) error {
			return reader.ReadMsg(historyRPCRequest)
		})
		if err != nil {
			log.Error("reading request", zap.Error(err))
			metrics.RecordStoreError(s.ctx, "decodeRPCFailure")
			return err
		}
		log = log.With(zap.String("id", historyRPCRequest.RequestId))
		if query := historyRPCRequest.Query; query != nil {
			log = log.With(logging.Filters(query.GetContentFilters()))
		}
		log.Info("received query")

		historyResponseRPC := &pb.HistoryRPC{}
		historyResponseRPC.RequestId = historyRPCRequest.RequestId
		var res *pb.HistoryResponse
		err = tracing.Wrap(ctx, "finding messages", func(ctx context.Context, span tracing.Span) (err error) {
			tracing.SpanType(span, "db")
			res, err = s.FindMessages(historyRPCRequest.Query)
			return err
		})
		if err != nil {
			log.Error("retrieving messages from DB", zap.Error(err))
			metrics.RecordStoreError(s.ctx, "dbError")
			return err
		}

		historyResponseRPC.Response = res
		log = log.With(
			zap.Int("messages", len(res.Messages)),
			logging.IfDebug(logging.PagingInfo(res.PagingInfo)),
		)
		tracing.SpanTag(span, "messages", len(res.Messages))

		err = tracing.Wrap(ctx, "writing response", func(ctx context.Context, span tracing.Span) error {
			return writer.WriteMsg(historyResponseRPC)
		})
		if err != nil {
			log.Error("writing response", zap.Error(err))
			_ = stream.Reset()
			return err
		}
		log.Info("response sent")
		return nil
	})
}

func (s *XmtpStore) storeIncomingMessages(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case envelope := <-s.MsgC:
			if envelope == nil {
				return
			}
			_, _ = s.storeMessage(envelope)
		}
	}
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
			metrics.EmitStoredMessages(ctx, s.db, s.log)
		}
	}
}

func (s *XmtpStore) storeMessage(env *protocol.Envelope) (stored bool, err error) {
	if isRelayPing(env) {
		s.log.Debug("not storing relay ping message")
		return false, nil
	}
	err = tracing.Wrap(s.ctx, "storing message", func(ctx context.Context, span tracing.Span) error {
		tracing.SpanResource(span, "store")
		tracing.SpanType(span, "p2p")
		err = s.msgProvider.Put(env) // Should the index be stored?
		if err != nil {
			tracing.SpanTag(span, "stored", false)
			if err, ok := err.(pgdriver.Error); ok && err.IntegrityViolation() {
				s.log.Debug("storing message", zap.Error(err))
				metrics.RecordStoreError(s.ctx, "store_duplicate_key")
				return nil
			}
			s.log.Error("storing message", zap.Error(err))
			metrics.RecordStoreError(s.ctx, "store_failure")
			span.Finish(tracing.WithError(err))
			return err
		}
		stored = true
		s.log.Info("message stored",
			zap.String("content_topic", env.Message().ContentTopic),
			zap.Int("size", env.Size()),
			logging.Time("sent", env.Index().SenderTime))
		// This expects me to know the length of the message queue, which I don't now that the store lives in the DB. Setting to 1 for now
		metrics.RecordMessage(s.ctx, "stored", 1)
		tracing.SpanTag(span, "stored", true)
		tracing.SpanTag(span, "content_topic", env.Message().ContentTopic)
		tracing.SpanTag(span, "size", env.Size())
		return nil
	})
	return stored, err
}

func isRelayPing(env *protocol.Envelope) bool {
	if env == nil || env.Message() == nil {
		return false
	}
	return env.Message().ContentTopic == relayPingContentTopic
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

// Selects multiple peers with store protocol instead of the default of just 1
func selectPeers(host host.Host, protocolId string, maxPeers int, log *zap.Logger) ([]peer.ID, error) {
	var peers peer.IDSlice
	for _, peer := range host.Peerstore().Peers() {
		protocols, err := host.Peerstore().SupportsProtocols(peer, protocolId)
		if err != nil {
			log.Error("error obtaining the protocols supported by peers", zap.Error(err))
			return nil, err
		}

		if len(protocols) > 0 {
			peers = append(peers, peer)
		}

		if len(peers) == maxPeers {
			break
		}
	}

	if len(peers) >= 1 {
		return peers, nil
	}

	return nil, utils.ErrNoPeersAvailable
}
