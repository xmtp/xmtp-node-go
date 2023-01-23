package store

import (
	"context"
	"testing"

	"github.com/status-im/go-waku/logging"
	"github.com/status-im/go-waku/waku/v2/utils"
	"github.com/stretchr/testify/require"
	test "github.com/xmtp/xmtp-node-go/pkg/testing"
)

// func TestStore_ExcludeRelayPings(t *testing.T) {
// 	t.Parallel()

// 	s, cleanup := newTestStore(t)
// 	defer cleanup()

// 	c := newTestClient(t, s.host.ID())
// 	addStoreProtocol(t, c.host, s.host)

// 	pubSubTopic := newTopic()

// 	storeMessage(t, s, test.NewMessage("topic1", 1, "msg1"), pubSubTopic)
// 	storeMessage(t, s, test.NewMessage(relayPingContentTopic, 2, ""), pubSubTopic)
// 	storeMessage(t, s, test.NewMessage("topic2", 3, "msg2"), pubSubTopic)
// 	storeMessage(t, s, test.NewMessage(relayPingContentTopic, 4, "msg4"), pubSubTopic)

// 	query := &pb.HistoryQuery{
// 		PubsubTopic: pubSubTopic,
// 	}
// 	expectQueryMessagesEventually(t, c, query, []*pb.WakuMessage{
// 		test.NewMessage("topic1", 1, "msg1"),
// 		test.NewMessage("topic2", 3, "msg2"),
// 	})
// }

func newTestStore(t *testing.T, opts ...Option) (*XmtpStore, func()) {
	db, _, dbCleanup := test.NewDB(t)
	log := utils.Logger()

	host := test.NewPeer(t)
	log = log.With(logging.HostID("node", host.ID()))
	store, err := NewXmtpStore(
		append(
			[]Option{
				WithLog(log),
				WithDB(db),
				WithReaderDB(db),
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

// func addStoreProtocol(t *testing.T, h1, h2 host.Host) {
// 	h1.Peerstore().AddAddr(h2.ID(), tests.GetHostAddress(h2), peerstore.PermanentAddrTTL)
// 	err := h1.Peerstore().AddProtocols(h2.ID(), string(store.StoreID_v20beta4))
// 	require.NoError(t, err)
// }

// func expectMessages(t *testing.T, s *XmtpStore, pubSubTopic string, msgs []*pb.WakuMessage) {
// 	res, err := s.FindMessages(&pb.HistoryQuery{
// 		PubsubTopic: pubSubTopic,
// 	})
// 	require.NoError(t, err)
// 	require.Empty(t, res.Error)
// 	require.Len(t, res.Messages, len(msgs))
// 	require.ElementsMatch(t, msgs, res.Messages)
// }

// func storeMessage(t *testing.T, s *XmtpStore, env *messagev1.Envelope, pubSubTopic string) {
// 	_, err := s.storeMessage(test.NewEnvelope(t, env, pubSubTopic))
// 	require.NoError(t, err)
// }

// func storeMessageWithTime(t *testing.T, s *XmtpStore, env *messagev1.Envelope, pubSubTopic string, receiverTime int64) {
// 	env := protocol.NewEnvelope(msg, receiverTime, pubSubTopic)
// 	_, err := s.storeMessage(env)
// 	require.NoError(t, err)
// }

// func newTopic() string {
// 	return "test-" + test.RandomStringLower(5)
// }
