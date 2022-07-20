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
	"github.com/xmtp/xmtp-node-go/pkg/logging"
	test "github.com/xmtp/xmtp-node-go/pkg/testing"
)

func TestStore_ExcludeRelayPings(t *testing.T) {
	t.Parallel()

	s, cleanup := newTestStore(t)
	defer cleanup()

	c := newTestClient(t, s.host.ID())
	addStoreProtocol(t, c.host, s.host)

	pubSubTopic := test.NewTopic()

	storeMessage(t, s, test.NewMessage("topic1", 1, "msg1"), pubSubTopic)
	storeMessage(t, s, test.NewMessage(relayPingContentTopic, 2, ""), pubSubTopic)
	storeMessage(t, s, test.NewMessage("topic2", 3, "msg2"), pubSubTopic)
	storeMessage(t, s, test.NewMessage(relayPingContentTopic, 4, "msg4"), pubSubTopic)

	query := &pb.HistoryQuery{
		PubsubTopic: pubSubTopic,
	}
	expectQueryMessagesEventually(t, c, query, []*pb.WakuMessage{
		test.NewMessage("topic1", 1, "msg1"),
		test.NewMessage("topic2", 3, "msg2"),
	})
}

func newTestStore(t *testing.T, opts ...Option) (*XmtpStore, func()) {
	db, _, dbCleanup := test.NewDB(t)
	log := utils.Logger()
	dbStore, err := NewDBStore(log, WithDBStoreDB(db))
	require.NoError(t, err)

	host := test.NewPeer(t)
	log = log.With(logging.HostID("node", host.ID()))
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
	_, err := s.storeMessage(test.NewEnvelope(t, msg, pubSubTopic))
	require.NoError(t, err)
}
