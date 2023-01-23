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
	"github.com/status-im/go-waku/waku/v2/utils"
	"github.com/stretchr/testify/require"
	"github.com/xmtp/xmtp-node-go/pkg/store"
	test "github.com/xmtp/xmtp-node-go/pkg/testing"
)

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
		store.WithDB(db),
		store.WithReaderDB(db),
	)
	require.NoError(t, err)

	store.Start(context.Background())

	return store, dbStore, store.Stop, dbCleanup
}

func storeResume(t *testing.T, n *wakunode.WakuNode) {
	_, err := n.Store().Resume(context.Background(), relay.DefaultWakuTopic, nil)
	require.NoError(t, err)
}

func listMessages(t *testing.T, s *store.XmtpStore, contentTopics []string) []*pb.WakuMessage {
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

func expectStoreMessagesEventually(t *testing.T, s *store.XmtpStore, contentTopics []string, expectedMsgs []*pb.WakuMessage) {
	msgs := listMessages(t, s, contentTopics)
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
			msgs = listMessages(t, s, contentTopics)
			if len(msgs) == len(expectedMsgs) {
				done = true
			}
		case <-timer:
			done = true
		}
	}
	require.ElementsMatch(t, expectedMsgs, msgs)
}

func newTopic() string {
	// A cleaner is enabled in the Store by default at this level, so we need
	// to namespace the topic to avoid it being deleted and flaking tests.
	return "/xmtp/test-" + test.RandomStringLower(5)
}
