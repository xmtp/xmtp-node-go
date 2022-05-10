package store

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"math"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-msgio/protoio"
	"github.com/status-im/go-waku/waku/v2/metrics"
	"github.com/status-im/go-waku/waku/v2/protocol"
	"github.com/status-im/go-waku/waku/v2/protocol/pb"
	"github.com/status-im/go-waku/waku/v2/protocol/store"
	"github.com/status-im/go-waku/waku/v2/utils"
	"github.com/xmtp/xmtp-node-go/logging"
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

	started bool

	msgProvider store.MessageProvider
	h           host.Host
}

func NewXmtpStore(host host.Host, db *sql.DB, p store.MessageProvider, maxRetentionDuration time.Duration, log *zap.Logger) *XmtpStore {
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
	xmtpStore.log = log.Named("store")
	xmtpStore.MsgC = make(chan *protocol.Envelope, bufferSize)

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

	s.wg.Add(1)
	go s.storeIncomingMessages(ctx)
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
	s.log.Error("Query not implemented")

	return nil, errors.New("XmtpStore.Query not implemented!")
}

// Next is used to retrieve the next page of rows from a query response.
// If no more records are found, the result will not contain any messages.
// This function is useful for iterating over results without having to manually
// specify the cursor and pagination order and max number of results
func (s *XmtpStore) Next(ctx context.Context, r *store.Result) (*store.Result, error) {
	s.log.Error("Next is not implemented")

	return nil, errors.New("Not implemented")
}

// Resume retrieves the history of waku messages published on the default waku pubsub topic since the last time the waku store node has been online
// messages are stored in the store node's messages field and in the message db
// the offline time window is measured as the difference between the current time and the timestamp of the most recent persisted waku message
// an offset of 20 second is added to the time window to count for nodes asynchrony
// the history is fetched from one of the peers persisted in the waku store node's peer manager unit
// peerList indicates the list of peers to query from. The history is fetched from the first available peer in this list. Such candidates should be found through a discovery method (to be developed).
// if no peerList is passed, one of the peers in the underlying peer manager unit of the store protocol is picked randomly to fetch the history from. The history gets fetched successfully if the dialed peer has been online during the queried time window.
// the resume proc returns the number of retrieved messages if no error occurs, otherwise returns the error string
func (s *XmtpStore) Resume(ctx context.Context, pubsubTopic string, peerList []peer.ID) (int, error) {
	if !s.started {
		return 0, errors.New("can't resume: store has not started")
	}

	currentTime := utils.GetUnixEpoch()
	lastSeenTime, err := s.findLastSeen()
	if err != nil {
		return 0, err
	}

	var offset int64 = int64(20 * time.Nanosecond)
	currentTime = currentTime + offset
	lastSeenTime = max(lastSeenTime-offset, 0)

	rpc := &pb.HistoryQuery{
		PubsubTopic: pubsubTopic,
		StartTime:   lastSeenTime,
		EndTime:     currentTime,
		PagingInfo: &pb.PagingInfo{
			PageSize:  0,
			Direction: pb.PagingInfo_BACKWARD,
		},
	}

	if len(peerList) == 0 {
		peerList, err = selectPeers(s.h, string(store.StoreID_v20beta4), maxPeersToResume, s.log)
		if err != nil {
			s.log.Error("Error selecting peer: ", zap.Error(err))
			return -1, store.ErrNoPeersAvailable
		}
	}

	messages, err := s.queryLoop(ctx, rpc, peerList)
	if err != nil {
		s.log.Error("resuming history", zap.Error(err))
		return -1, store.ErrFailedToResumeHistory
	}

	msgCount := 0
	for _, msg := range messages {
		if err = s.storeMessage(protocol.NewEnvelope(msg, pubsubTopic)); err == nil {
			msgCount++
		}
	}

	s.log.Info("Retrieved messages since the last online time", zap.Int("count", msgCount))

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

func (s *XmtpStore) queryFrom(ctx context.Context, q *pb.HistoryQuery, selectedPeer peer.ID, requestId []byte) (*pb.HistoryResponse, error) {
	s.log.Info("querying message history", logging.HostID("peer", selectedPeer))

	// We connect first so dns4 addresses are resolved (NewStream does not do it)
	err := s.h.Connect(ctx, s.h.Peerstore().PeerInfo(selectedPeer))
	if err != nil {
		return nil, err
	}

	connOpt, err := s.h.NewStream(ctx, selectedPeer, store.StoreID_v20beta4)
	if err != nil {
		s.log.Error("connecting to peer", zap.Error(err))
		return nil, err
	}

	defer connOpt.Close()
	defer func() {
		_ = connOpt.Reset()
	}()

	historyRequest := &pb.HistoryRPC{Query: q, RequestId: hex.EncodeToString(requestId)}

	writer := protoio.NewDelimitedWriter(connOpt)
	reader := protoio.NewDelimitedReader(connOpt, math.MaxInt32)

	err = writer.WriteMsg(historyRequest)
	if err != nil {
		s.log.Error("writing request", zap.Error(err))
		return nil, err
	}

	historyResponseRPC := &pb.HistoryRPC{}
	err = reader.ReadMsg(historyResponseRPC)
	if err != nil {
		s.log.Error("reading response", zap.Error(err))
		metrics.RecordStoreError(s.ctx, "decodeRPCFailure")
		return nil, err
	}

	metrics.RecordMessage(ctx, "retrieved", 1)

	return historyResponseRPC.Response, nil
}

func (s *XmtpStore) queryLoop(ctx context.Context, query *pb.HistoryQuery, candidateList []peer.ID) ([]*pb.WakuMessage, error) {
	// loops through the candidateList in order and sends the query to each until one of the query gets resolved successfully
	// returns the number of retrieved messages, or error if all the requests fail

	queryWg := sync.WaitGroup{}
	queryWg.Add(len(candidateList))

	resultChan := make(chan *pb.HistoryResponse, len(candidateList))

	for _, candidate := range candidateList {
		go func(peer peer.ID) {
			defer queryWg.Done()
			result, err := s.queryFrom(ctx, query, peer, protocol.GenerateRequestId())
			if err == nil {
				resultChan <- result
				return
			}
			s.log.Error("resuming history", logging.HostID("peer", peer), zap.Error(err))
		}(candidate)
	}

	queryWg.Wait()
	close(resultChan)

	var messages []*pb.WakuMessage
	hasResults := false
	for result := range resultChan {
		hasResults = true
		messages = append(messages, result.Messages...)
	}

	if hasResults {
		return messages, nil
	}

	return nil, store.ErrFailedQuery
}

func (s *XmtpStore) onRequest(stream network.Stream) {
	defer stream.Close()

	historyRPCRequest := &pb.HistoryRPC{}

	writer := protoio.NewDelimitedWriter(stream)
	reader := protoio.NewDelimitedReader(stream, math.MaxInt32)

	err := reader.ReadMsg(historyRPCRequest)
	if err != nil {
		s.log.Error("reading request", zap.Error(err))
		metrics.RecordStoreError(s.ctx, "decodeRPCFailure")
		return
	}

	s.log.Info("received query", logging.HostID("peer", stream.Conn().RemotePeer()))

	historyResponseRPC := &pb.HistoryRPC{}
	historyResponseRPC.RequestId = historyRPCRequest.RequestId
	res, err := s.FindMessages(historyRPCRequest.Query)
	if err != nil {
		s.log.Error("Error retrieving messages from DB")
		metrics.RecordStoreError(s.ctx, "dbError")
		return
	}

	historyResponseRPC.Response = res

	err = writer.WriteMsg(historyResponseRPC)
	if err != nil {
		s.log.Error("writing response", zap.Error(err))
		_ = stream.Reset()
	} else {
		s.log.Info("response sent", logging.HostID("peer", stream.Conn().RemotePeer()))
	}
}

func (s *XmtpStore) storeIncomingMessages(ctx context.Context) {
	defer s.wg.Done()
	for envelope := range s.MsgC {
		_ = s.storeMessage(envelope)
	}
}

func (s *XmtpStore) storeMessage(env *protocol.Envelope) error {
	index, err := computeIndex(env)
	if err != nil {
		s.log.Error("creating message index", zap.Error(err))
		return err
	}
	err = s.msgProvider.Put(index, env.PubsubTopic(), env.Message()) // Should the index be stored?
	if err != nil {
		s.log.Error("storing message", zap.Error(err))
		metrics.RecordStoreError(s.ctx, "store_failure")
		return err
	}
	// This expects me to know the length of the message queue, which I don't now that the store lives in the DB. Setting to 1 for now
	metrics.RecordMessage(s.ctx, "stored", 1)

	return nil
}

func computeIndex(env *protocol.Envelope) (*pb.Index, error) {
	return &pb.Index{
		Digest:       env.Hash(),
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
func selectPeers(host host.Host, protocolId string, maxPeers int, log *zap.SugaredLogger) ([]peer.ID, error) {
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
