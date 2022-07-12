package store

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	libp2ptest "github.com/libp2p/go-libp2p-core/test"
	"github.com/status-im/go-waku/waku/v2/protocol"
	"github.com/status-im/go-waku/waku/v2/protocol/pb"
	"github.com/stretchr/testify/require"
	test "github.com/xmtp/xmtp-node-go/pkg/testing"
)

func TestStoreClient_New(t *testing.T) {
	host := test.NewPeer(t)
	log := test.NewLog(t)
	peerID := libp2ptest.RandPeerIDFatal(t)
	tcs := []struct {
		name   string
		opts   []ClientOption
		expect func(t *testing.T, c *Client, err error)
	}{
		{
			name: "missing log",
			opts: []ClientOption{},
			expect: func(t *testing.T, c *Client, err error) {
				require.Equal(t, ErrMissingLogOption, err)
				require.Nil(t, c)
			},
		},
		{
			name: "missing host",
			opts: []ClientOption{
				WithClientLog(log),
			},
			expect: func(t *testing.T, c *Client, err error) {
				require.Equal(t, ErrMissingHostOption, err)
				require.Nil(t, c)
			},
		},
		{
			name: "missing peer",
			opts: []ClientOption{
				WithClientLog(log),
				WithClientHost(host),
			},
			expect: func(t *testing.T, c *Client, err error) {
				require.Equal(t, ErrMissingPeerOption, err)
				require.Nil(t, c)
			},
		},
		{
			name: "success",
			opts: []ClientOption{
				WithClientLog(log),
				WithClientHost(host),
				WithClientPeer(peerID),
			},
			expect: func(t *testing.T, c *Client, err error) {
				require.NoError(t, err)
				require.NotNil(t, c)
			},
		},
	}
	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c, err := NewClient(tc.opts...)
			tc.expect(t, c, err)
		})
	}
}

func TestStoreClient_Query(t *testing.T) {
	tcs := []struct {
		name     string
		query    *pb.HistoryQuery
		stored   []*protocol.Envelope
		expected []*pb.WakuMessage
	}{
		{
			name:     "empty no messages",
			query:    &pb.HistoryQuery{},
			stored:   []*protocol.Envelope{},
			expected: []*pb.WakuMessage{},
		},
		{
			name:  "empty with messages",
			query: &pb.HistoryQuery{},
			stored: []*protocol.Envelope{
				test.NewEnvelope(t, test.NewMessage("topic1", 1, "msg1"), "pubsub1"),
				test.NewEnvelope(t, test.NewMessage("topic2", 2, "msg2"), "pubsub2"),
			},
			expected: []*pb.WakuMessage{
				test.NewMessage("topic1", 1, "msg1"),
				test.NewMessage("topic2", 2, "msg2"),
			},
		},
		{
			name: "with content topic",
			query: &pb.HistoryQuery{
				ContentFilters: buildContentFilters("topic1"),
			},
			stored: []*protocol.Envelope{
				test.NewEnvelope(t, test.NewMessage("topic1", 1, "msg1"), "pubsub1"),
				test.NewEnvelope(t, test.NewMessage("topic2", 2, "msg2"), "pubsub2"),
			},
			expected: []*pb.WakuMessage{
				test.NewMessage("topic1", 1, "msg1"),
			},
		},
		{
			name: "with multiple content topics",
			query: &pb.HistoryQuery{
				ContentFilters: buildContentFilters("topic1", "topic3"),
			},
			stored: []*protocol.Envelope{
				test.NewEnvelope(t, test.NewMessage("topic1", 1, "msg1"), "pubsub1"),
				test.NewEnvelope(t, test.NewMessage("topic2", 2, "msg2"), "pubsub1"),
				test.NewEnvelope(t, test.NewMessage("topic3", 3, "msg3"), "pubsub1"),
				test.NewEnvelope(t, test.NewMessage("topic4", 4, "msg4"), "pubsub1"),
			},
			expected: []*pb.WakuMessage{
				test.NewMessage("topic1", 1, "msg1"),
				test.NewMessage("topic3", 3, "msg3"),
			},
		},
		{
			name: "with pubsub topic",
			query: &pb.HistoryQuery{
				PubsubTopic: "pubsub2",
			},
			stored: []*protocol.Envelope{
				test.NewEnvelope(t, test.NewMessage("topic1", 1, "msg1"), "pubsub1"),
				test.NewEnvelope(t, test.NewMessage("topic2", 2, "msg2"), "pubsub2"),
			},
			expected: []*pb.WakuMessage{
				test.NewMessage("topic2", 2, "msg2"),
			},
		},
		{
			name: "with start time",
			query: &pb.HistoryQuery{
				StartTime: 2,
			},
			stored: []*protocol.Envelope{
				test.NewEnvelope(t, test.NewMessage("topic1", 1, "msg1"), "pubsub1"),
				test.NewEnvelope(t, test.NewMessage("topic2", 2, "msg2"), "pubsub2"),
				test.NewEnvelope(t, test.NewMessage("topic1", 3, "msg3"), "pubsub3"),
				test.NewEnvelope(t, test.NewMessage("topic2", 4, "msg4"), "pubsub4"),
			},
			expected: []*pb.WakuMessage{
				test.NewMessage("topic2", 2, "msg2"),
				test.NewMessage("topic1", 3, "msg3"),
				test.NewMessage("topic2", 4, "msg4"),
			},
		},
		{
			name: "with end time",
			query: &pb.HistoryQuery{
				EndTime: 3,
			},
			stored: []*protocol.Envelope{
				test.NewEnvelope(t, test.NewMessage("topic1", 1, "msg1"), "pubsub1"),
				test.NewEnvelope(t, test.NewMessage("topic2", 2, "msg2"), "pubsub2"),
				test.NewEnvelope(t, test.NewMessage("topic1", 3, "msg3"), "pubsub3"),
				test.NewEnvelope(t, test.NewMessage("topic2", 4, "msg4"), "pubsub4"),
			},
			expected: []*pb.WakuMessage{
				test.NewMessage("topic1", 1, "msg1"),
				test.NewMessage("topic2", 2, "msg2"),
				test.NewMessage("topic1", 3, "msg3"),
			},
		},
		{
			name: "with start and end time",
			query: &pb.HistoryQuery{
				StartTime: 2,
				EndTime:   3,
			},
			stored: []*protocol.Envelope{
				test.NewEnvelope(t, test.NewMessage("topic1", 1, "msg1"), "pubsub1"),
				test.NewEnvelope(t, test.NewMessage("topic2", 2, "msg2"), "pubsub2"),
				test.NewEnvelope(t, test.NewMessage("topic1", 3, "msg3"), "pubsub3"),
				test.NewEnvelope(t, test.NewMessage("topic2", 4, "msg4"), "pubsub4"),
			},
			expected: []*pb.WakuMessage{
				test.NewMessage("topic2", 2, "msg2"),
				test.NewMessage("topic1", 3, "msg3"),
			},
		},
		{
			name: "with content topic start and end time",
			query: &pb.HistoryQuery{
				ContentFilters: buildContentFilters("topic1"),
				StartTime:      2,
				EndTime:        3,
			},
			stored: []*protocol.Envelope{
				test.NewEnvelope(t, test.NewMessage("topic1", 1, "msg1"), "pubsub1"),
				test.NewEnvelope(t, test.NewMessage("topic2", 2, "msg2"), "pubsub2"),
				test.NewEnvelope(t, test.NewMessage("topic1", 3, "msg3"), "pubsub3"),
				test.NewEnvelope(t, test.NewMessage("topic2", 4, "msg4"), "pubsub4"),
			},
			expected: []*pb.WakuMessage{
				test.NewMessage("topic1", 3, "msg3"),
			},
		},
		{
			name: "with pubsub topic content topic start and end time",
			query: &pb.HistoryQuery{
				PubsubTopic:    "pubsub2",
				ContentFilters: buildContentFilters("topic2"),
				StartTime:      2,
				EndTime:        5,
			},
			stored: []*protocol.Envelope{
				test.NewEnvelope(t, test.NewMessage("topic1", 1, "msg1"), "pubsub1"),
				test.NewEnvelope(t, test.NewMessage("topic2", 2, "msg2"), "pubsub2"),
				test.NewEnvelope(t, test.NewMessage("topic1", 3, "msg3"), "pubsub1"),
				test.NewEnvelope(t, test.NewMessage("topic2", 4, "msg4"), "pubsub2"),
				test.NewEnvelope(t, test.NewMessage("topic2", 5, "msg5"), "pubsub1"),
				test.NewEnvelope(t, test.NewMessage("topic2", 6, "msg4"), "pubsub2"),
			},
			expected: []*pb.WakuMessage{
				test.NewMessage("topic2", 2, "msg2"),
				test.NewMessage("topic2", 4, "msg4"),
			},
		},
		{
			name: "with pubsub topic and content topic",
			query: &pb.HistoryQuery{
				PubsubTopic:    "pubsub2",
				ContentFilters: buildContentFilters("topic2"),
			},
			stored: []*protocol.Envelope{
				test.NewEnvelope(t, test.NewMessage("topic1", 1, "msg1"), "pubsub1"),
				test.NewEnvelope(t, test.NewMessage("topic2", 2, "msg2"), "pubsub2"),
				test.NewEnvelope(t, test.NewMessage("topic1", 3, "msg3"), "pubsub1"),
				test.NewEnvelope(t, test.NewMessage("topic2", 4, "msg4"), "pubsub2"),
				test.NewEnvelope(t, test.NewMessage("topic2", 5, "msg5"), "pubsub1"),
				test.NewEnvelope(t, test.NewMessage("topic2", 6, "msg6"), "pubsub2"),
			},
			expected: []*pb.WakuMessage{
				test.NewMessage("topic2", 2, "msg2"),
				test.NewMessage("topic2", 4, "msg4"),
				test.NewMessage("topic2", 6, "msg6"),
			},
		},
		{
			name: "with paging page size",
			query: &pb.HistoryQuery{
				PubsubTopic:    "pubsub1",
				ContentFilters: buildContentFilters("topic1", "topic2"),
				PagingInfo: &pb.PagingInfo{
					PageSize: 2,
				},
			},
			stored: []*protocol.Envelope{
				test.NewEnvelope(t, test.NewMessage("topic1", 1, "msg1"), "pubsub1"),
				test.NewEnvelope(t, test.NewMessage("topic2", 2, "msg2"), "pubsub1"),
				test.NewEnvelope(t, test.NewMessage("topic1", 3, "msg3"), "pubsub1"),
				test.NewEnvelope(t, test.NewMessage("topic2", 4, "msg4"), "pubsub1"),
				test.NewEnvelope(t, test.NewMessage("topic3", 5, "msg5"), "pubsub1"),
				test.NewEnvelope(t, test.NewMessage("topic2", 6, "msg6"), "pubsub1"),
				test.NewEnvelope(t, test.NewMessage("topic1", 7, "msg7"), "pubsub2"),
			},
			expected: []*pb.WakuMessage{
				test.NewMessage("topic1", 1, "msg1"),
				test.NewMessage("topic2", 2, "msg2"),
				test.NewMessage("topic1", 3, "msg3"),
				test.NewMessage("topic2", 4, "msg4"),
				test.NewMessage("topic2", 6, "msg6"),
			},
		},
		{
			name: "with paging cursor forward",
			query: &pb.HistoryQuery{
				PubsubTopic:    "pubsub1",
				ContentFilters: buildContentFilters("topic1", "topic2"),
				PagingInfo: &pb.PagingInfo{
					PageSize:  2,
					Cursor:    test.NewEnvelope(t, test.NewMessage("topic1", 3, "msg3"), "pubsub1").Index(),
					Direction: pb.PagingInfo_FORWARD,
				},
			},
			stored: []*protocol.Envelope{
				test.NewEnvelope(t, test.NewMessage("topic1", 1, "msg1"), "pubsub1"),
				test.NewEnvelope(t, test.NewMessage("topic2", 2, "msg2"), "pubsub1"),
				test.NewEnvelope(t, test.NewMessage("topic1", 3, "msg3"), "pubsub1"),
				test.NewEnvelope(t, test.NewMessage("topic2", 4, "msg4"), "pubsub1"),
				test.NewEnvelope(t, test.NewMessage("topic3", 5, "msg5"), "pubsub1"),
				test.NewEnvelope(t, test.NewMessage("topic2", 6, "msg6"), "pubsub1"),
				test.NewEnvelope(t, test.NewMessage("topic1", 7, "msg7"), "pubsub2"),
			},
			expected: []*pb.WakuMessage{
				test.NewMessage("topic2", 4, "msg4"),
				test.NewMessage("topic2", 6, "msg6"),
			},
		},
		{
			name: "with paging cursor backward",
			query: &pb.HistoryQuery{
				PubsubTopic:    "pubsub1",
				ContentFilters: buildContentFilters("topic1", "topic2"),
				PagingInfo: &pb.PagingInfo{
					PageSize:  1,
					Cursor:    test.NewEnvelope(t, test.NewMessage("topic1", 3, "msg3"), "pubsub1").Index(),
					Direction: pb.PagingInfo_BACKWARD,
				},
			},
			stored: []*protocol.Envelope{
				test.NewEnvelope(t, test.NewMessage("topic1", 1, "msg1"), "pubsub1"),
				test.NewEnvelope(t, test.NewMessage("topic2", 2, "msg2"), "pubsub1"),
				test.NewEnvelope(t, test.NewMessage("topic1", 3, "msg3"), "pubsub1"),
				test.NewEnvelope(t, test.NewMessage("topic2", 4, "msg4"), "pubsub1"),
				test.NewEnvelope(t, test.NewMessage("topic3", 5, "msg5"), "pubsub1"),
				test.NewEnvelope(t, test.NewMessage("topic2", 6, "msg6"), "pubsub1"),
				test.NewEnvelope(t, test.NewMessage("topic1", 7, "msg7"), "pubsub2"),
			},
			expected: []*pb.WakuMessage{
				test.NewMessage("topic1", 1, "msg1"),
				test.NewMessage("topic2", 2, "msg2"),
			},
		},
	}
	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			s, cleanup := newTestStore(t)
			defer cleanup()
			c := newTestClient(t, s.h.ID())
			addStoreProtocol(t, c.host, s.h)
			for _, env := range tc.stored {
				storeMessage(t, s, env.Message(), env.PubsubTopic())
			}
			expectQueryMessagesEventually(t, c, tc.query, tc.expected)
		})
	}
}

func TestStoreClient_Query_PagingShouldStopOnReturnFalse(t *testing.T) {
	s, cleanup := newTestStore(t)
	defer cleanup()
	c := newTestClient(t, s.h.ID())
	addStoreProtocol(t, c.host, s.h)
	storeMessage(t, s, test.NewMessage("topic1", 1, "msg1"), "pubsub1")
	storeMessage(t, s, test.NewMessage("topic1", 2, "msg2"), "pubsub1")
	storeMessage(t, s, test.NewMessage("topic1", 3, "msg3"), "pubsub1")
	storeMessage(t, s, test.NewMessage("topic1", 4, "msg4"), "pubsub1")
	var page int
	ctx := context.Background()
	msgCount, err := c.Query(ctx, &pb.HistoryQuery{
		PagingInfo: &pb.PagingInfo{
			PageSize: 1,
		},
	}, func(res *pb.HistoryResponse) (int, bool) {
		page++
		if page == 3 {
			return 0, false
		}
		return len(res.Messages), true
	})
	require.NoError(t, err)
	require.Equal(t, 2, msgCount)
}

func newTestClient(t *testing.T, peerID peer.ID) *Client {
	host := test.NewPeer(t)
	log := test.NewLog(t)
	c, err := NewClient(
		WithClientLog(log),
		WithClientHost(host),
		WithClientPeer(peerID),
	)
	require.NoError(t, err)
	require.NotNil(t, c)
	return c
}

func expectQueryMessagesEventually(t *testing.T, c *Client, query *pb.HistoryQuery, expectedMsgs []*pb.WakuMessage) []*pb.WakuMessage {
	var msgs []*pb.WakuMessage
	require.Eventually(t, func() bool {
		msgs = queryMessages(t, c, query)
		return len(msgs) == len(expectedMsgs)
	}, 3*time.Second, 500*time.Millisecond, "expected %d == %d", len(msgs), len(expectedMsgs))
	require.ElementsMatch(t, expectedMsgs, msgs)
	return msgs
}

func queryMessages(t *testing.T, client *Client, query *pb.HistoryQuery) []*pb.WakuMessage {
	var msgs []*pb.WakuMessage
	ctx := context.Background()
	msgCount, err := client.Query(ctx, query, func(res *pb.HistoryResponse) (int, bool) {
		msgs = append(msgs, res.Messages...)
		return len(res.Messages), true
	})
	require.NoError(t, err)
	require.Equal(t, msgCount, len(msgs))
	return msgs
}

func buildContentFilters(contentTopics ...string) []*pb.ContentFilter {
	contentFilters := make([]*pb.ContentFilter, len(contentTopics))
	for i, contentTopic := range contentTopics {
		contentFilters[i] = &pb.ContentFilter{
			ContentTopic: contentTopic,
		}
	}
	return contentFilters
}
