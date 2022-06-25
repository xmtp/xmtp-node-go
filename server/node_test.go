package server

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/status-im/go-waku/waku/v2/node"
	wakunode "github.com/status-im/go-waku/waku/v2/node"
	"github.com/status-im/go-waku/waku/v2/protocol/pb"
	"github.com/status-im/go-waku/waku/v2/protocol/relay"
	wakustore "github.com/status-im/go-waku/waku/v2/protocol/store"
	"github.com/status-im/go-waku/waku/v2/utils"
	"github.com/stretchr/testify/require"
	"github.com/xmtp/xmtp-node-go/store"
	test "github.com/xmtp/xmtp-node-go/testing"
)

func TestNode_Resume_OnStart_StoreNodesConnectedBefore(t *testing.T) {
	n1, cleanup := newTestNode(t, nil, false)
	defer cleanup()

	topic1 := test.NewTopic()
	topic2 := test.NewTopic()

	test.Publish(t, n1, topic1, 1)
	test.Publish(t, n1, topic2, 2)

	n2, cleanup := newTestNode(t, []*node.WakuNode{n1}, true)
	defer cleanup()

	expectStoreMessagesEventually(t, n2, []string{topic1, topic2}, []*pb.WakuMessage{
		test.NewMessage(topic1, 1),
		test.NewMessage(topic2, 2),
	})
}

func TestNode_Resume_OnStart_StoreNodesConnectedAfter(t *testing.T) {
	n1, cleanup := newTestNode(t, nil, false)
	defer cleanup()

	topic1 := test.NewTopic()
	topic2 := test.NewTopic()

	test.Publish(t, n1, topic1, 1)
	test.Publish(t, n1, topic2, 2)

	n2, cleanup := newTestNode(t, nil, true)
	defer cleanup()
	test.ConnectStoreNode(t, n2, n1)

	expectStoreMessagesEventually(t, n2, []string{topic1, topic2}, []*pb.WakuMessage{
		test.NewMessage(topic1, 1),
		test.NewMessage(topic2, 2),
	})
}
func TestNode_DataPartition_WithoutResume(t *testing.T) {
	n1, cleanup := newTestNode(t, nil, false)
	defer cleanup()

	n2, cleanup := newTestNode(t, nil, false)
	defer cleanup()

	// Connect and send a message to each node, expecting that the messages
	// are relayed to the other nodes.
	test.Connect(t, n1, n2)

	n1EnvC := test.Subscribe(t, n1)
	n2EnvC := test.Subscribe(t, n2)

	topic1 := test.NewTopic()
	topic2 := test.NewTopic()

	test.Publish(t, n1, topic1, 1)
	test.Publish(t, n2, topic2, 2)

	test.SubscribeExpect(t, n1EnvC, []*pb.WakuMessage{
		test.NewMessage(topic1, 1),
		test.NewMessage(topic2, 2),
	})
	test.SubscribeExpect(t, n2EnvC, []*pb.WakuMessage{
		test.NewMessage(topic1, 1),
		test.NewMessage(topic2, 2),
	})
	test.SubscribeExpectNone(t, n1EnvC)
	test.SubscribeExpectNone(t, n2EnvC)

	expectStoreMessages(t, n1, []string{topic1, topic2}, []*pb.WakuMessage{
		test.NewMessage(topic1, 1),
		test.NewMessage(topic2, 2),
	})
	expectStoreMessages(t, n2, []string{topic1, topic2}, []*pb.WakuMessage{
		test.NewMessage(topic1, 1),
		test.NewMessage(topic2, 2),
	})

	// Disconnect and send a message to each node, expecting that the messages
	// are not relayed to the other node.
	test.Disconnect(t, n1, n2)

	test.Publish(t, n1, topic1, 4)
	test.Publish(t, n2, topic2, 5)
	test.Publish(t, n1, topic1, 6)
	test.Publish(t, n2, topic2, 7)

	test.SubscribeExpect(t, n1EnvC, []*pb.WakuMessage{
		test.NewMessage(topic1, 4),
		test.NewMessage(topic1, 6),
	})
	test.SubscribeExpect(t, n2EnvC, []*pb.WakuMessage{
		test.NewMessage(topic2, 5),
		test.NewMessage(topic2, 7),
	})
	test.SubscribeExpectNone(t, n1EnvC)
	test.SubscribeExpectNone(t, n2EnvC)

	expectStoreMessages(t, n1, []string{topic1, topic2}, []*pb.WakuMessage{
		test.NewMessage(topic1, 1),
		test.NewMessage(topic2, 2),
		test.NewMessage(topic1, 4),
		test.NewMessage(topic1, 6),
	})
	expectStoreMessages(t, n2, []string{topic1, topic2}, []*pb.WakuMessage{
		test.NewMessage(topic1, 1),
		test.NewMessage(topic2, 2),
		test.NewMessage(topic2, 5),
		test.NewMessage(topic2, 7),
	})

	// Reconnect and expect that no new messages are relayed.
	test.Connect(t, n1, n2)

	test.SubscribeExpectNone(t, n1EnvC)
	test.SubscribeExpectNone(t, n2EnvC)

	expectStoreMessages(t, n1, []string{topic1, topic2}, []*pb.WakuMessage{
		test.NewMessage(topic1, 1),
		test.NewMessage(topic2, 2),
		test.NewMessage(topic1, 4),
		test.NewMessage(topic1, 6),
	})
	expectStoreMessages(t, n2, []string{topic1, topic2}, []*pb.WakuMessage{
		test.NewMessage(topic1, 1),
		test.NewMessage(topic2, 2),
		test.NewMessage(topic2, 5),
		test.NewMessage(topic2, 7),
	})
}

func TestNode_DataPartition_WithResume(t *testing.T) {
	n1, cleanup := newTestNode(t, nil, true)
	defer cleanup()

	n2, cleanup := newTestNode(t, nil, true)
	defer cleanup()

	// Connect and send a message to each node, expecting that the messages
	// are relayed to the other nodes.
	test.Connect(t, n1, n2)

	n1EnvC := test.Subscribe(t, n1)
	n2EnvC := test.Subscribe(t, n2)

	topic1 := test.NewTopic()
	topic2 := test.NewTopic()

	test.Publish(t, n1, topic1, 1)
	test.Publish(t, n2, topic2, 2)

	test.SubscribeExpect(t, n1EnvC, []*pb.WakuMessage{
		test.NewMessage(topic1, 1),
		test.NewMessage(topic2, 2),
	})
	test.SubscribeExpect(t, n2EnvC, []*pb.WakuMessage{
		test.NewMessage(topic1, 1),
		test.NewMessage(topic2, 2),
	})
	test.SubscribeExpectNone(t, n1EnvC)
	test.SubscribeExpectNone(t, n2EnvC)

	expectStoreMessages(t, n1, []string{topic1, topic2}, []*pb.WakuMessage{
		test.NewMessage(topic1, 1),
		test.NewMessage(topic2, 2),
	})
	expectStoreMessages(t, n2, []string{topic1, topic2}, []*pb.WakuMessage{
		test.NewMessage(topic1, 1),
		test.NewMessage(topic2, 2),
	})

	// Disconnect and send a message to each node, expecting that the messages
	// are not relayed to the other node.
	test.Disconnect(t, n1, n2)

	test.Publish(t, n1, topic1, 4)
	test.Publish(t, n2, topic2, 5)
	test.Publish(t, n1, topic1, 6)
	test.Publish(t, n2, topic2, 7)

	test.SubscribeExpect(t, n1EnvC, []*pb.WakuMessage{
		test.NewMessage(topic1, 4),
		test.NewMessage(topic1, 6),
	})
	test.SubscribeExpect(t, n2EnvC, []*pb.WakuMessage{
		test.NewMessage(topic2, 5),
		test.NewMessage(topic2, 7),
	})
	test.SubscribeExpectNone(t, n1EnvC)
	test.SubscribeExpectNone(t, n2EnvC)

	expectStoreMessages(t, n1, []string{topic1, topic2}, []*pb.WakuMessage{
		test.NewMessage(topic1, 1),
		test.NewMessage(topic2, 2),
		test.NewMessage(topic1, 4),
		test.NewMessage(topic1, 6),
	})
	expectStoreMessages(t, n2, []string{topic1, topic2}, []*pb.WakuMessage{
		test.NewMessage(topic1, 1),
		test.NewMessage(topic2, 2),
		test.NewMessage(topic2, 5),
		test.NewMessage(topic2, 7),
	})

	// Reconnect, trigger a resume from node 2, and expect new messages .
	test.ConnectStoreNode(t, n1, n2)
	msgCount, err := n1.Store().Resume(context.Background(), relay.DefaultWakuTopic, []peer.ID{n2.Host().ID()})
	require.NoError(t, err)
	require.Equal(t, 2, msgCount)

	test.SubscribeExpectNone(t, n1EnvC)
	test.SubscribeExpectNone(t, n2EnvC)

	expectStoreMessages(t, n1, []string{topic1, topic2}, []*pb.WakuMessage{
		test.NewMessage(topic1, 1),
		test.NewMessage(topic2, 2),
		test.NewMessage(topic1, 4),
		test.NewMessage(topic2, 5),
		test.NewMessage(topic1, 6),
		test.NewMessage(topic2, 7),
	})
	expectStoreMessages(t, n2, []string{topic1, topic2}, []*pb.WakuMessage{
		test.NewMessage(topic1, 1),
		test.NewMessage(topic2, 2),
		test.NewMessage(topic2, 5),
		test.NewMessage(topic2, 7),
	})
}

func newTestNode(t *testing.T, storeNodes []*wakunode.WakuNode, withResume bool) (*wakunode.WakuNode, func()) {
	return test.NewNode(t, storeNodes,
		wakunode.WithWakuStore(true, withResume),
		wakunode.WithWakuStoreFactory(func(w *wakunode.WakuNode) wakustore.Store {
			// Note that the same host needs to be used here.
			store, _, _ := newTestStore(t, w.Host())
			return store
		}),
	)
}

func newTestStore(t *testing.T, host host.Host) (*store.XmtpStore, *store.DBStore, func()) {
	db, dbCleanup := test.NewDB(t)
	dbStore, err := store.NewDBStore(utils.Logger(), store.WithDB(db))
	require.NoError(t, err)

	if host == nil {
		host = test.NewPeer(t)
	}
	store := store.NewXmtpStore(host, db, dbStore, 0, utils.Logger())

	store.Start(context.Background())

	return store, dbStore, func() {
		store.Stop()
		dbCleanup()
	}
}

func listMessages(t *testing.T, n *wakunode.WakuNode, contentTopics []string) []*pb.WakuMessage {
	s := n.Store().(*store.XmtpStore)
	contentFilters := make([]*pb.ContentFilter, len(contentTopics))
	for i, contentTopic := range contentTopics {
		contentFilters[i] = &pb.ContentFilter{
			ContentTopic: contentTopic,
		}
	}
	res, err := s.FindMessages(&pb.HistoryQuery{
		PubsubTopic:    relay.DefaultWakuTopic,
		ContentFilters: contentFilters,
	})
	require.NoError(t, err)
	return res.Messages
}

func expectStoreMessagesEventually(t *testing.T, n *node.WakuNode, contentTopics []string, expectedMsgs []*pb.WakuMessage) {
	var msgs []*pb.WakuMessage
	require.Eventually(t, func() bool {
		msgs = listMessages(t, n, contentTopics)
		return len(msgs) == 2
	}, 3*time.Second, 100*time.Millisecond)
	require.ElementsMatch(t, expectedMsgs, msgs)
}

func expectStoreMessages(t *testing.T, n *node.WakuNode, contentTopics []string, expectedMsgs []*pb.WakuMessage) {
	msgs := listMessages(t, n, contentTopics)
	require.ElementsMatch(t, expectedMsgs, msgs)
}
