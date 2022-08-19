package server

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p-core/host"

	wakunode "github.com/status-im/go-waku/waku/v2/node"
	"github.com/status-im/go-waku/waku/v2/protocol/pb"
	"github.com/status-im/go-waku/waku/v2/protocol/relay"
	wakustore "github.com/status-im/go-waku/waku/v2/protocol/store"
	"github.com/status-im/go-waku/waku/v2/utils"
	"github.com/stretchr/testify/require"
	"github.com/xmtp/xmtp-node-go/pkg/store"
	test "github.com/xmtp/xmtp-node-go/pkg/testing"
)

func TestNode_PublishSubscribeQuery_DifferentDBs(t *testing.T) {
	t.Parallel()

	n1, cleanup := newTestNode(t, nil, false, nil)
	defer cleanup()

	n2, cleanup := newTestNode(t, nil, false, nil)
	defer cleanup()

	topic1 := test.NewTopic()
	topic2 := test.NewTopic()

	// Connect to each other as store nodes.
	test.Connect(t, n1, n2, string(wakustore.StoreID_v20beta4))
	test.Connect(t, n2, n1, string(wakustore.StoreID_v20beta4))
	test.ExpectPeers(t, n1, n2.Host().ID())
	test.ExpectPeers(t, n2, n1.Host().ID())

	// Subscribe via each node.
	n1EnvC := test.Subscribe(t, n1)
	n2EnvC := test.Subscribe(t, n2)

	// Publish to each node.
	test.Publish(t, n1, test.NewMessage(topic1, 1, "msg1"))
	test.Publish(t, n2, test.NewMessage(topic2, 2, "msg2"))

	// Expect subscribed messages.
	expectedMsgs := []*pb.WakuMessage{
		test.NewMessage(topic1, 1, "msg1"),
		test.NewMessage(topic2, 2, "msg2"),
	}
	test.SubscribeExpect(t, n1EnvC, expectedMsgs)
	test.SubscribeExpect(t, n2EnvC, expectedMsgs)

	// Expect query messages.
	expectStoreMessagesEventually(t, n1, []string{topic1, topic2}, expectedMsgs)
}

func TestNode_PublishSubscribeQuery_SharedDB(t *testing.T) {
	t.Parallel()

	db, _, cleanup := test.NewDB(t)
	defer cleanup()

	n1, cleanup := newTestNode(t, nil, false, db)
	defer cleanup()

	n2, cleanup := newTestNode(t, nil, false, db)
	defer cleanup()

	topic1 := test.NewTopic()
	topic2 := test.NewTopic()

	// Connect to each other as store nodes.
	test.Connect(t, n1, n2, string(wakustore.StoreID_v20beta4))
	test.Connect(t, n2, n1, string(wakustore.StoreID_v20beta4))
	test.ExpectPeers(t, n1, n2.Host().ID())
	test.ExpectPeers(t, n2, n1.Host().ID())

	// Subscribe via each node.
	n1EnvC := test.Subscribe(t, n1)
	n2EnvC := test.Subscribe(t, n2)

	// Publish to each node.
	test.Publish(t, n1, test.NewMessage(topic1, 1, "msg1"))
	test.Publish(t, n2, test.NewMessage(topic2, 2, "msg2"))

	// Expect subscribed messages.
	expectedMsgs := []*pb.WakuMessage{
		test.NewMessage(topic1, 1, "msg1"),
		test.NewMessage(topic2, 2, "msg2"),
	}
	test.SubscribeExpect(t, n1EnvC, expectedMsgs)
	test.SubscribeExpect(t, n2EnvC, expectedMsgs)

	// Expect query messages.
	expectStoreMessagesEventually(t, n1, []string{topic1, topic2}, expectedMsgs)
}

func TestNode_Resume_OnStart_StoreNodesConnectedBefore(t *testing.T) {
	t.Parallel()

	n1, cleanup := newTestNode(t, nil, false, nil)
	defer cleanup()

	topic1 := test.NewTopic()
	topic2 := test.NewTopic()

	test.Publish(t, n1, test.NewMessage(topic1, 1, "msg1"))
	test.Publish(t, n1, test.NewMessage(topic2, 2, "msg2"))

	n2, cleanup := newTestNode(t, []*wakunode.WakuNode{n1}, true, nil)
	defer cleanup()

	expectStoreMessagesEventually(t, n2, []string{topic1, topic2}, []*pb.WakuMessage{
		test.NewMessage(topic1, 1, "msg1"),
		test.NewMessage(topic2, 2, "msg2"),
	})
}

func TestNode_Resume_OnStart_StoreNodesConnectedAfter(t *testing.T) {
	t.Parallel()

	n1, cleanup := newTestNode(t, nil, false, nil)
	defer cleanup()

	topic1 := test.NewTopic()
	topic2 := test.NewTopic()

	test.Publish(t, n1, test.NewMessage(topic1, 1, "msg1"))
	test.Publish(t, n1, test.NewMessage(topic2, 2, "msg2"))

	n2, cleanup := newTestNode(t, nil, true, nil)
	defer cleanup()
	test.ConnectStoreNode(t, n2, n1)

	expectStoreMessagesEventually(t, n2, []string{topic1, topic2}, []*pb.WakuMessage{
		test.NewMessage(topic1, 1, "msg1"),
		test.NewMessage(topic2, 2, "msg2"),
	})
}
func TestNode_DataPartition_WithoutResume(t *testing.T) {
	t.Parallel()

	n1, cleanup := newTestNode(t, nil, false, nil)
	defer cleanup()

	n2, cleanup := newTestNode(t, nil, false, nil)
	defer cleanup()

	// Connect and send a message to each node, expecting that the messages
	// are relayed to the other nodes.
	test.Connect(t, n1, n2)
	test.ExpectPeers(t, n1, n2.Host().ID())
	test.ExpectPeers(t, n2, n1.Host().ID())

	topic1 := test.NewTopic()
	topic2 := test.NewTopic()

	n1EnvC := test.Subscribe(t, n1)
	n2EnvC := test.Subscribe(t, n2)

	test.Publish(t, n1, test.NewMessage(topic1, 1, "msg1"))
	test.Publish(t, n2, test.NewMessage(topic2, 2, "msg2"))

	test.SubscribeExpect(t, n1EnvC, []*pb.WakuMessage{
		test.NewMessage(topic1, 1, "msg1"),
		test.NewMessage(topic2, 2, "msg2"),
	})
	test.SubscribeExpect(t, n2EnvC, []*pb.WakuMessage{
		test.NewMessage(topic1, 1, "msg1"),
		test.NewMessage(topic2, 2, "msg2"),
	})
	test.SubscribeExpectNone(t, n1EnvC)
	test.SubscribeExpectNone(t, n2EnvC)

	expectStoreMessagesEventually(t, n1, []string{topic1, topic2}, []*pb.WakuMessage{
		test.NewMessage(topic1, 1, "msg1"),
		test.NewMessage(topic2, 2, "msg2"),
	})
	expectStoreMessagesEventually(t, n2, []string{topic1, topic2}, []*pb.WakuMessage{
		test.NewMessage(topic1, 1, "msg1"),
		test.NewMessage(topic2, 2, "msg2"),
	})

	// Disconnect and send a message to each node, expecting that the messages
	// are not relayed to the other node.
	test.Disconnect(t, n1, n2)
	test.ExpectNoPeers(t, n1)
	test.ExpectNoPeers(t, n2)

	test.Publish(t, n1, test.NewMessage(topic1, 4, "msg4"))
	test.Publish(t, n2, test.NewMessage(topic2, 5, "msg5"))
	test.Publish(t, n1, test.NewMessage(topic1, 6, "msg6"))
	test.Publish(t, n2, test.NewMessage(topic2, 7, "msg7"))

	test.SubscribeExpect(t, n1EnvC, []*pb.WakuMessage{
		test.NewMessage(topic1, 4, "msg4"),
		test.NewMessage(topic1, 6, "msg6"),
	})
	test.SubscribeExpect(t, n2EnvC, []*pb.WakuMessage{
		test.NewMessage(topic2, 5, "msg5"),
		test.NewMessage(topic2, 7, "msg7"),
	})
	test.SubscribeExpectNone(t, n1EnvC)
	test.SubscribeExpectNone(t, n2EnvC)

	expectStoreMessagesEventually(t, n1, []string{topic1, topic2}, []*pb.WakuMessage{
		test.NewMessage(topic1, 1, "msg1"),
		test.NewMessage(topic2, 2, "msg2"),
		test.NewMessage(topic1, 4, "msg4"),
		test.NewMessage(topic1, 6, "msg6"),
	})
	expectStoreMessagesEventually(t, n2, []string{topic1, topic2}, []*pb.WakuMessage{
		test.NewMessage(topic1, 1, "msg1"),
		test.NewMessage(topic2, 2, "msg2"),
		test.NewMessage(topic2, 5, "msg5"),
		test.NewMessage(topic2, 7, "msg7"),
	})

	// Reconnect and expect that no new messages are relayed.
	test.Connect(t, n1, n2)
	test.ExpectPeers(t, n1, n2.Host().ID())
	test.ExpectPeers(t, n2, n1.Host().ID())

	test.SubscribeExpectNone(t, n1EnvC)
	test.SubscribeExpectNone(t, n2EnvC)

	expectStoreMessagesEventually(t, n1, []string{topic1, topic2}, []*pb.WakuMessage{
		test.NewMessage(topic1, 1, "msg1"),
		test.NewMessage(topic2, 2, "msg2"),
		test.NewMessage(topic1, 4, "msg4"),
		test.NewMessage(topic1, 6, "msg6"),
	})
	expectStoreMessagesEventually(t, n2, []string{topic1, topic2}, []*pb.WakuMessage{
		test.NewMessage(topic1, 1, "msg1"),
		test.NewMessage(topic2, 2, "msg2"),
		test.NewMessage(topic2, 5, "msg5"),
		test.NewMessage(topic2, 7, "msg7"),
	})
}

func TestNode_DataPartition_WithResume(t *testing.T) {
	t.Parallel()

	n1, cleanup := newTestNode(t, nil, true, nil)
	defer cleanup()

	n2, cleanup := newTestNode(t, nil, true, nil)
	defer cleanup()

	// Connect and send a message to each node, expecting that the messages
	// are relayed to the other nodes.
	test.Connect(t, n1, n2)
	test.ExpectPeers(t, n1, n2.Host().ID())
	test.ExpectPeers(t, n2, n1.Host().ID())

	topic1 := test.NewTopic()
	topic2 := test.NewTopic()

	n1EnvC := test.Subscribe(t, n1)
	n2EnvC := test.Subscribe(t, n2)

	test.Publish(t, n1, test.NewMessage(topic1, 1, "msg1"))
	test.Publish(t, n2, test.NewMessage(topic2, 2, "msg2"))

	test.SubscribeExpect(t, n1EnvC, []*pb.WakuMessage{
		test.NewMessage(topic1, 1, "msg1"),
		test.NewMessage(topic2, 2, "msg2"),
	})
	test.SubscribeExpect(t, n2EnvC, []*pb.WakuMessage{
		test.NewMessage(topic1, 1, "msg1"),
		test.NewMessage(topic2, 2, "msg2"),
	})
	test.SubscribeExpectNone(t, n1EnvC)
	test.SubscribeExpectNone(t, n2EnvC)

	expectStoreMessagesEventually(t, n1, []string{topic1, topic2}, []*pb.WakuMessage{
		test.NewMessage(topic1, 1, "msg1"),
		test.NewMessage(topic2, 2, "msg2"),
	})
	expectStoreMessagesEventually(t, n2, []string{topic1, topic2}, []*pb.WakuMessage{
		test.NewMessage(topic1, 1, "msg1"),
		test.NewMessage(topic2, 2, "msg2"),
	})

	// Disconnect and send a message to each node, expecting that the messages
	// are not relayed to the other node.
	test.Disconnect(t, n1, n2)
	test.ExpectNoPeers(t, n1)
	test.ExpectNoPeers(t, n2)

	test.Publish(t, n1, test.NewMessage(topic1, 4, "msg4"))
	test.Publish(t, n2, test.NewMessage(topic2, 5, "msg5"))
	test.Publish(t, n1, test.NewMessage(topic1, 6, "msg6"))
	test.Publish(t, n2, test.NewMessage(topic2, 7, "msg7"))

	test.SubscribeExpect(t, n1EnvC, []*pb.WakuMessage{
		test.NewMessage(topic1, 4, "msg4"),
		test.NewMessage(topic1, 6, "msg6"),
	})
	test.SubscribeExpect(t, n2EnvC, []*pb.WakuMessage{
		test.NewMessage(topic2, 5, "msg5"),
		test.NewMessage(topic2, 7, "msg7"),
	})
	test.SubscribeExpectNone(t, n1EnvC)
	test.SubscribeExpectNone(t, n2EnvC)

	expectStoreMessagesEventually(t, n1, []string{topic1, topic2}, []*pb.WakuMessage{
		test.NewMessage(topic1, 1, "msg1"),
		test.NewMessage(topic2, 2, "msg2"),
		test.NewMessage(topic1, 4, "msg4"),
		test.NewMessage(topic1, 6, "msg6"),
	})
	expectStoreMessagesEventually(t, n2, []string{topic1, topic2}, []*pb.WakuMessage{
		test.NewMessage(topic1, 1, "msg1"),
		test.NewMessage(topic2, 2, "msg2"),
		test.NewMessage(topic2, 5, "msg5"),
		test.NewMessage(topic2, 7, "msg7"),
	})

	// Reconnect, resume is automatically triggered from node 2, and expect new
	// messages.
	test.ConnectStoreNode(t, n1, n2)
	test.ExpectPeers(t, n1, n2.Host().ID())
	test.ExpectPeers(t, n2, n1.Host().ID())

	test.SubscribeExpectNone(t, n1EnvC)
	test.SubscribeExpectNone(t, n2EnvC)

	expectStoreMessagesEventually(t, n1, []string{topic1, topic2}, []*pb.WakuMessage{
		test.NewMessage(topic1, 1, "msg1"),
		test.NewMessage(topic2, 2, "msg2"),
		test.NewMessage(topic1, 4, "msg4"),
		test.NewMessage(topic2, 5, "msg5"),
		test.NewMessage(topic1, 6, "msg6"),
		test.NewMessage(topic2, 7, "msg7"),
	})
	expectStoreMessagesEventually(t, n2, []string{topic1, topic2}, []*pb.WakuMessage{
		test.NewMessage(topic1, 1, "msg1"),
		test.NewMessage(topic2, 2, "msg2"),
		test.NewMessage(topic1, 4, "msg4"),
		test.NewMessage(topic2, 5, "msg5"),
		test.NewMessage(topic1, 6, "msg6"),
		test.NewMessage(topic2, 7, "msg7"),
	})
}

func TestNodes_Deployment(t *testing.T) {
	tcs := []struct {
		name string
		test func(t *testing.T, n1, newN1, n2, newN2 *wakunode.WakuNode)
	}{
		{
			name: "new instances connect to new instances",
			test: func(t *testing.T, n1, newN1, n2, newN2 *wakunode.WakuNode) {
				n1ID := n1.Host().ID()
				n2ID := n2.Host().ID()

				// Deploy started; new instances connect to all new instances.
				test.Connect(t, newN1, newN2)
				test.Connect(t, newN2, newN1)

				// Expect new instances are fully connected.
				test.ExpectPeers(t, newN1, n2ID)
				test.ExpectPeers(t, newN2, n1ID)

				// Deploy ending; old instances disconnect from connected peers.
				test.Disconnect(t, n1, n2)

				// Expect new instances are still fully connected.
				test.ExpectPeers(t, newN1, n2ID)
				test.ExpectPeers(t, newN2, n1ID)
			},
		},
		{
			name: "new instances connect to 1 new 1 old instance",
			test: func(t *testing.T, n1, newN1, n2, newN2 *wakunode.WakuNode) {
				n1ID := n1.Host().ID()
				n2ID := n2.Host().ID()

				// Deploy started; new instances connect to some new instances
				// and some old instances.
				test.Connect(t, newN1, newN2)
				test.Connect(t, newN2, n1)

				// Expect new instances are fully connected.
				test.ExpectPeers(t, newN1, n2ID)
				test.ExpectPeers(t, newN2, n1ID)

				// Deploy ending; old instances disconnect from connected peers.
				test.Disconnect(t, n1, n2)

				// Expect new instances are still fully connected.
				test.ExpectPeers(t, newN1, n2ID)
				test.ExpectPeers(t, newN2, n1ID)
			},
		},
		{
			name: "new instances connect to old instances",
			test: func(t *testing.T, n1, newN1, n2, newN2 *wakunode.WakuNode) {
				n1ID := n1.Host().ID()
				n2ID := n2.Host().ID()

				// Deploy started; new instances connect to all old instances.
				test.Connect(t, newN1, n2)
				test.Connect(t, newN2, n1)

				// Expect new instances are fully connected.
				test.ExpectPeers(t, newN1, n2ID)
				test.ExpectPeers(t, newN2, n1ID)

				// Deploy ending; old instances disconnect from connected peers.
				test.Disconnect(t, n1, n2)

				// Expect new instances are fully disconnected.
				test.ExpectNoPeers(t, newN1)
				test.ExpectNoPeers(t, newN2)
			},
		},
	}
	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Initialize private keys for each node.
			n1PrivKey := test.NewPrivateKey(t)
			n2PrivKey := test.NewPrivateKey(t)

			// Spin up initial instances of the nodes.
			n1, cleanup := newTestNode(t, nil, false, nil, wakunode.WithPrivateKey(n1PrivKey))
			defer cleanup()
			n2, cleanup := newTestNode(t, nil, false, nil, wakunode.WithPrivateKey(n2PrivKey))
			defer cleanup()

			// Connect the nodes.
			test.Connect(t, n1, n2)
			test.Connect(t, n2, n1)

			// Spin up new instances of the nodes.
			newN1, cleanup := newTestNode(t, nil, false, nil, wakunode.WithPrivateKey(n1PrivKey))
			defer cleanup()
			newN2, cleanup := newTestNode(t, nil, false, nil, wakunode.WithPrivateKey(n2PrivKey))
			defer cleanup()

			// Expect matching peer IDs for new and old instances.
			require.Equal(t, n1.Host().ID(), newN1.Host().ID())
			require.Equal(t, n2.Host().ID(), newN2.Host().ID())

			// Run the test case.
			tc.test(t, n1, newN1, n2, newN2)
		})
	}
}

func newTestNode(t *testing.T, storeNodes []*wakunode.WakuNode, withResume bool, db *sql.DB, opts ...wakunode.WakuNodeOption) (*wakunode.WakuNode, func()) {
	var dbCleanup func()
	n, nodeCleanup := test.NewNode(t, storeNodes,
		append(
			opts,
			wakunode.WithWakuStore(true, withResume),
			wakunode.WithWakuStoreFactory(func(w *wakunode.WakuNode) wakustore.Store {
				// Note that the node calls store.Stop() during it's cleanup,
				// but it that doesn't clean up the given DB, so we make sure
				// to return that in the node cleanup returned here.
				// Note that the same host needs to be used here.
				var store *store.XmtpStore
				store, _, _, dbCleanup = newTestStore(t, w.Host(), db)
				return store
			}),
		)...,
	)
	return n, func() {
		nodeCleanup()
		if dbCleanup != nil {
			dbCleanup()
		}
	}
}

func newTestStore(t *testing.T, host host.Host, db *sql.DB) (*store.XmtpStore, *store.DBStore, func(), func()) {
	var dbCleanup func()
	if db == nil {
		db, _, dbCleanup = test.NewDB(t)
	}
	dbStore, err := store.NewDBStore(utils.Logger(), store.WithDBStoreDB(db))
	require.NoError(t, err)

	if host == nil {
		host = test.NewPeer(t)
	}
	store, err := store.NewXmtpStore(
		store.WithLog(utils.Logger()),
		store.WithHost(host),
		store.WithDB(db),
		store.WithMessageProvider(dbStore),
	)
	require.NoError(t, err)

	store.Start(context.Background())

	return store, dbStore, store.Stop, dbCleanup
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

func expectStoreMessagesEventually(t *testing.T, n *wakunode.WakuNode, contentTopics []string, expectedMsgs []*pb.WakuMessage) {

	msgs := listMessages(t, n, contentTopics)
	if len(msgs) == len(expectedMsgs) {
		require.ElementsMatch(t, expectedMsgs, msgs)
		return
	}

	timer := time.After(3 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	var done bool
	for !done {
		select {
		case <-ticker.C:
			msgs = listMessages(t, n, contentTopics)
			if len(msgs) == len(expectedMsgs) {
				done = true
			}
		case <-timer:
			done = true
		}
	}
	require.ElementsMatch(t, expectedMsgs, msgs)
}
