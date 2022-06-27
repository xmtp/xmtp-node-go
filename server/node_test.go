package server

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
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
