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

type XmtpStore struct {
	ctx  context.Context
	MsgC chan *protocol.Envelope
	wg   *sync.WaitGroup
	db   *sql.DB

	log *zap.Logger

	started        bool
	statsPeriod    time.Duration
	resumePageSize int

	msgProvider store.MessageProvider
	h           host.Host
}

func NewXmtpStore(host host.Host, db *sql.DB, p store.MessageProvider, statsPeriod time.Duration, log *zap.Logger) *XmtpStore {
	if log == nil {
		panic("No logger provided in store initialization")
	}
	if db == nil {
		log.Panic("No DB in store initialization")
	}
	if p == nil {
		log.Panic("No message provider in store initialization")
	}
	xmtpStore := new(XmtpStore)
	xmtpStore.msgProvider = p
	xmtpStore.h = host
	xmtpStore.db = db
	xmtpStore.wg = &sync.WaitGroup{}
	xmtpStore.log = log.Named("store").With(zap.String("node", host.ID().Pretty()))
	xmtpStore.MsgC = make(chan *protocol.Envelope, bufferSize)
	xmtpStore.statsPeriod = statsPeriod

	return xmtpStore
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
	s.h.SetStreamHandler(store.StoreID_v20beta4, s.onRequest)

	s.wg.Add(2)
	go tracing.Do(ctx, "store-incoming-messages", func(ctx context.Context) { s.storeIncomingMessages(ctx) })
	go tracing.Do(ctx, "store-status-metrics", func(ctx context.Context) { s.statusMetricsLoop(ctx) })
	s.log.Info("Store protocol started")
}

func (s *XmtpStore) Stop() {
	s.started = false

	if s.MsgC != nil {
		close(s.MsgC)
	}

	if s.h != nil {
		s.h.RemoveStreamHandler(store.StoreID_v20beta4)
	}

	s.wg.Wait()
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
	log.Info("resuming")

	currentTime := utils.GetUnixEpoch()
	lastSeenTime, err := s.findLastSeen()
	if err != nil {
		return 0, errors.Wrap(err, "finding latest timestamp")
	}

	var offset int64 = int64(20 * time.Second)
	currentTime = currentTime + offset
	lastSeenTime = max(lastSeenTime-offset, 0)

	if len(peers) == 0 {
		peers, err = selectPeers(s.h, string(store.StoreID_v20beta4), maxPeersToResume, s.log)
		if err != nil {
			return 0, errors.Wrap(err, "selecting peer")
		}
	}
	if len(peers) == 0 {
		return 0, errors.New("no peers")
	}

	req := &pb.HistoryQuery{
		PubsubTopic: pubsubTopic,
		StartTime:   lastSeenTime,
		EndTime:     currentTime,
		PagingInfo: &pb.PagingInfo{
			PageSize:  uint64(s.resumePageSize),
			Direction: pb.PagingInfo_BACKWARD,
		},
	}
	var wg sync.WaitGroup
	var msgCount int
	var msgCountLock sync.RWMutex
	for _, p := range peers {
		wg.Add(1)
		go func(p peer.ID) {
			defer wg.Done()
			count, err := s.queryPeer(ctx, req, p, func(msg *pb.WakuMessage) bool {
				err, stored := s.storeMessage(protocol.NewEnvelope(msg, utils.GetUnixEpoch(), req.PubsubTopic))
				if err != nil {
					s.log.Error("storing message", zap.Error(err))
					return false
				}
				return stored
			})
			if err != nil {
				s.log.Error("querying peer", zap.Error(err), zap.String("peer", p.Pretty()))
				return
			}
			msgCountLock.Lock()
			defer msgCountLock.Unlock()
			msgCount += count
		}(p)
	}
	wg.Wait()
	s.log.Info("resume complete", zap.Int("count", msgCount))

	return msgCount, nil
}

func (s *XmtpStore) queryPeer(ctx context.Context, req *pb.HistoryQuery, peerID peer.ID, msgFn func(*pb.WakuMessage) bool) (int, error) {
	c, err := NewClient(
		WithClientLog(s.log),
		WithClientHost(s.h),
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
	span, _ := tracing.StartSpanFromContext(s.ctx, "store request")
	defer span.Finish()
	log := s.log.With(logging.HostID("peer", stream.Conn().RemotePeer()))
	log = tracing.Link(span, log)
	span.SetTag("peer", stream.Conn().RemotePeer())

	historyRPCRequest := &pb.HistoryRPC{}
	writer := protoio.NewDelimitedWriter(stream)
	reader := protoio.NewDelimitedReader(stream, math.MaxInt32)

	err := reader.ReadMsg(historyRPCRequest)
	if err != nil {
		log.Error("reading request", zap.Error(err))
		metrics.RecordStoreError(s.ctx, "decodeRPCFailure")
		span.Finish(tracing.WithError(err))
		return
	}
	log = log.With(zap.String("id", historyRPCRequest.RequestId))
	if query := historyRPCRequest.Query; query != nil {
		log = log.With(logging.Filters(query.GetContentFilters()))
	}
	log.Info("received query")

	historyResponseRPC := &pb.HistoryRPC{}
	historyResponseRPC.RequestId = historyRPCRequest.RequestId
	res, err := s.FindMessages(historyRPCRequest.Query)
	if err != nil {
		log.Error("retrieving messages from DB", zap.Error(err))
		metrics.RecordStoreError(s.ctx, "dbError")
		span.Finish(tracing.WithError(err))
		return
	}

	historyResponseRPC.Response = res
	log = log.With(
		zap.Int("messages", len(res.Messages)),
		logging.IfDebug(logging.PagingInfo(res.PagingInfo)),
	)
	span.SetTag("messages", len(res.Messages))

	err = writer.WriteMsg(historyResponseRPC)
	if err != nil {
		log.Error("writing response", zap.Error(err))
		_ = stream.Reset()
		span.Finish(tracing.WithError(err))
	} else {
		log.Info("response sent")
	}
}

func (s *XmtpStore) storeIncomingMessages(ctx context.Context) {
	defer s.wg.Done()
	for envelope := range s.MsgC {
		_, _ = s.storeMessage(envelope)
	}
}

func (s *XmtpStore) statusMetricsLoop(ctx context.Context) {
	defer s.wg.Done()
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

func (s *XmtpStore) storeMessage(env *protocol.Envelope) (error, bool) {
	span, _ := tracing.StartSpanFromContext(s.ctx, "store message")
	defer span.Finish()
	log := tracing.Link(span, s.log)
	err := s.msgProvider.Put(env) // Should the index be stored?
	if err != nil {
		if err, ok := err.(pgdriver.Error); ok && err.IntegrityViolation() {
			log.Debug("storing message", zap.Error(err))
			metrics.RecordStoreError(s.ctx, "store_duplicate_key")
			return nil, false
		}
		log.Error("storing message", zap.Error(err))
		metrics.RecordStoreError(s.ctx, "store_failure")
		span.Finish(tracing.WithError(err))
		return err, false
	}
	log.Info("message stored",
		zap.String("content_topic", env.Message().ContentTopic),
		zap.Int("size", env.Size()),
		logging.Time("sent", env.Index().SenderTime))
	// This expects me to know the length of the message queue, which I don't now that the store lives in the DB. Setting to 1 for now
	metrics.RecordMessage(s.ctx, "stored", 1)
	span.SetTag("content_topic", env.Message().ContentTopic)
	span.SetTag("size", env.Size())
	return nil, true
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
