package store

import (
	"context"
	"testing"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/status-im/go-waku/tests"
	"github.com/status-im/go-waku/waku/v2/protocol/pb"
	"github.com/status-im/go-waku/waku/v2/protocol/store"
	"github.com/status-im/go-waku/waku/v2/utils"
	"github.com/stretchr/testify/require"
	test "github.com/xmtp/xmtp-node-go/pkg/testing"
	"go.uber.org/zap"
)

func newTestStore(t *testing.T, opts ...Option) (*XmtpStore, func()) {
	db, _, dbCleanup := test.NewDB(t)
	log := utils.Logger()
	dbStore, err := NewDBStore(log, WithDBStoreDB(db))
	require.NoError(t, err)

	host := test.NewPeer(t)
	log = log.With(zap.String("node", host.ID().Pretty()))
	store, err := NewXmtpStore(
		append(
			[]Option{
				WithLog(log),
				WithHost(host),
				WithDB(db),
				WithMessageProvider(dbStore),
			},
			opts...,
		)...,
	)
	require.NoError(t, err)

	store.Start(context.Background())

	return store, func() {
		store.Stop()
		dbCleanup()
	}
}

func addStoreProtocol(t *testing.T, h1, h2 host.Host) {
	h1.Peerstore().AddAddr(h2.ID(), tests.GetHostAddress(h2), peerstore.PermanentAddrTTL)
	err := h1.Peerstore().AddProtocols(h2.ID(), string(store.StoreID_v20beta4))
	require.NoError(t, err)
}

func expectMessages(t *testing.T, s *XmtpStore, pubSubTopic string, msgs []*pb.WakuMessage) {
	res, err := s.FindMessages(&pb.HistoryQuery{
		PubsubTopic: pubSubTopic,
	})
	require.NoError(t, err)
	require.Empty(t, res.Error)
	require.Len(t, res.Messages, len(msgs))
	require.ElementsMatch(t, msgs, res.Messages)
}

func storeMessage(t *testing.T, s *XmtpStore, msg *pb.WakuMessage, pubSubTopic string) {
	err, _ := s.storeMessage(test.NewEnvelope(t, msg, pubSubTopic))
	require.NoError(t, err)
}
