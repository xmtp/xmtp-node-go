package store

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-msgio/protoio"
	"github.com/status-im/go-waku/waku/v2/metrics"
	"github.com/status-im/go-waku/waku/v2/protocol"
	"github.com/status-im/go-waku/waku/v2/protocol/pb"
	"github.com/status-im/go-waku/waku/v2/protocol/store"
	"github.com/status-im/go-waku/waku/v2/utils"
	"go.uber.org/zap"
)

const bufferSize = 1024

type XmtpStore struct {
	ctx  context.Context
	MsgC chan *protocol.Envelope
	wg   *sync.WaitGroup
	db   *sql.DB

	log *zap.SugaredLogger

	started bool

	msgProvider store.MessageProvider
	h           host.Host
}

func NewXmtpStore(host host.Host, db *sql.DB, p store.MessageProvider, maxRetentionDuration time.Duration, log *zap.SugaredLogger) *XmtpStore {
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
	s.h.SetStreamHandlerMatch(store.StoreID_v20beta4, protocol.PrefixTextMatch(string(store.StoreID_v20beta4)), s.onRequest)

	s.wg.Add(1)
	go s.storeIncomingMessages(ctx)
}

func (s *XmtpStore) FindMessages(query *pb.HistoryQuery) (res *pb.HistoryResponse, err error) {
	return FindMessages(s.db, query)
}

func (s *XmtpStore) onRequest(stream network.Stream) {
	defer stream.Close()

	historyRPCRequest := &pb.HistoryRPC{}

	writer := protoio.NewDelimitedWriter(stream)
	reader := protoio.NewDelimitedReader(stream, math.MaxInt32)

	err := reader.ReadMsg(historyRPCRequest)
	if err != nil {
		s.log.Error("error reading request", err)
		metrics.RecordStoreError(s.ctx, "decodeRPCFailure")
		return
	}

	s.log.Info(fmt.Sprintf("%s: Received query from %s", stream.Conn().LocalPeer(), stream.Conn().RemotePeer()))

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
		s.log.Error("error writing response", err)
		_ = stream.Reset()
	} else {
		s.log.Info(fmt.Sprintf("%s: Response sent  to %s", stream.Conn().LocalPeer().String(), stream.Conn().RemotePeer().String()))
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
		s.log.Error("could not calculate message index", err)
		return err
	}
	err = s.msgProvider.Put(index, env.PubsubTopic(), env.Message()) // Should the index be stored?
	if err != nil {
		s.log.Error("could not store message", err)
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
