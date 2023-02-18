package store

import (
	"testing"
	"time"

	"github.com/status-im/go-waku/waku/v2/protocol/pb"
	"github.com/stretchr/testify/require"
	test "github.com/xmtp/xmtp-node-go/pkg/testing"
)

func TestStore_Cleaner(t *testing.T) {
	pubSubTopic := "test-" + test.RandomStringLower(5)
	tcs := []struct {
		name      string
		batchSize int
		setup     func(t *testing.T, s *XmtpStore)
		expected  []*pb.WakuMessage
	}{
		{
			name:      "no messages",
			batchSize: 5,
			setup:     func(t *testing.T, s *XmtpStore) {},
			expected:  []*pb.WakuMessage{},
		},
		{
			name:      "less than batch size",
			batchSize: 5,
			setup: func(t *testing.T, s *XmtpStore) {
				fourDaysAgo := time.Now().UTC().Add(-4 * 24 * time.Hour).UnixNano()
				storeMessageWithTime(t, s, test.NewMessage("topic1", 1, "msg1"), pubSubTopic, fourDaysAgo)
				storeMessageWithTime(t, s, test.NewMessage("/xmtp/topic2", 2, "msg2"), pubSubTopic, fourDaysAgo)
				storeMessageWithTime(t, s, test.NewMessage("topic3", 3, "msg3"), pubSubTopic, fourDaysAgo)
				storeMessageWithTime(t, s, test.NewMessage("/xmtp/topic4", 4, "msg4"), pubSubTopic, fourDaysAgo)

				storeMessage(t, s, test.NewMessage("topic5", 5, "msg5"), pubSubTopic)
				storeMessage(t, s, test.NewMessage("/xmtp/topic6", 6, "msg6"), pubSubTopic)
				storeMessage(t, s, test.NewMessage("topic7", 7, "msg7"), pubSubTopic)
				storeMessage(t, s, test.NewMessage("/xmtp/topic8", 8, "msg8"), pubSubTopic)
			},
			expected: []*pb.WakuMessage{
				test.NewMessage("topic1", 1, "msg1"),
				test.NewMessage("/xmtp/topic2", 2, "msg2"),
				test.NewMessage("topic3", 3, "msg3"),
				test.NewMessage("/xmtp/topic4", 4, "msg4"),
				test.NewMessage("topic5", 5, "msg5"),
				test.NewMessage("/xmtp/topic6", 6, "msg6"),
				test.NewMessage("topic7", 7, "msg7"),
				test.NewMessage("/xmtp/topic8", 8, "msg8"),
			},
		},
		{
			name:      "equal to batch size",
			batchSize: 3,
			setup: func(t *testing.T, s *XmtpStore) {
				fourDaysAgo := time.Now().UTC().Add(-4 * 24 * time.Hour).UnixNano()
				storeMessageWithTime(t, s, test.NewMessage("topic1", 1, "msg1"), pubSubTopic, fourDaysAgo)
				storeMessageWithTime(t, s, test.NewMessage("/xmtp/topic2", 2, "msg2"), pubSubTopic, fourDaysAgo)
				storeMessageWithTime(t, s, test.NewMessage("topic3", 3, "msg3"), pubSubTopic, fourDaysAgo)
				storeMessageWithTime(t, s, test.NewMessage("/xmtp/topic4", 4, "msg4"), pubSubTopic, fourDaysAgo)
				storeMessageWithTime(t, s, test.NewMessage("topic5", 5, "msg5"), pubSubTopic, fourDaysAgo)
				storeMessageWithTime(t, s, test.NewMessage("/xmtp/topic6", 6, "msg6"), pubSubTopic, fourDaysAgo)

				storeMessage(t, s, test.NewMessage("topic7", 7, "msg7"), pubSubTopic)
				storeMessage(t, s, test.NewMessage("/xmtp/topic8", 8, "msg8"), pubSubTopic)
			},
			expected: []*pb.WakuMessage{
				test.NewMessage("/xmtp/topic2", 2, "msg2"),
				test.NewMessage("/xmtp/topic4", 4, "msg4"),
				test.NewMessage("/xmtp/topic6", 6, "msg6"),
				test.NewMessage("topic7", 7, "msg7"),
				test.NewMessage("/xmtp/topic8", 8, "msg8"),
			},
		},
		{
			name:      "greater than batch size",
			batchSize: 2,
			setup: func(t *testing.T, s *XmtpStore) {
				fourDaysAgo := time.Now().UTC().Add(-4 * 24 * time.Hour).UnixNano()
				storeMessageWithTime(t, s, test.NewMessage("topic1", 1, "msg1"), pubSubTopic, fourDaysAgo)
				storeMessageWithTime(t, s, test.NewMessage("/xmtp/topic2", 2, "msg2"), pubSubTopic, fourDaysAgo)
				storeMessageWithTime(t, s, test.NewMessage("topic3", 3, "msg3"), pubSubTopic, fourDaysAgo)
				storeMessageWithTime(t, s, test.NewMessage("/xmtp/topic4", 4, "msg4"), pubSubTopic, fourDaysAgo)
				storeMessageWithTime(t, s, test.NewMessage("topic5", 5, "msg5"), pubSubTopic, fourDaysAgo)
				storeMessageWithTime(t, s, test.NewMessage("/xmtp/topic6", 6, "msg6"), pubSubTopic, fourDaysAgo)

				storeMessage(t, s, test.NewMessage("topic7", 7, "msg7"), pubSubTopic)
				storeMessage(t, s, test.NewMessage("/xmtp/topic8", 8, "msg8"), pubSubTopic)
			},
			expected: []*pb.WakuMessage{
				test.NewMessage("/xmtp/topic2", 2, "msg2"),
				test.NewMessage("/xmtp/topic4", 4, "msg4"),
				test.NewMessage("topic5", 5, "msg5"),
				test.NewMessage("/xmtp/topic6", 6, "msg6"),
				test.NewMessage("topic7", 7, "msg7"),
				test.NewMessage("/xmtp/topic8", 8, "msg8"),
			},
		},
	}
	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s, cleanup := newTestStore(t, WithCleaner(CleanerOptions{
				Enable:        true,
				ActivePeriod:  time.Second,
				PassivePeriod: time.Second,
				RetentionDays: 1,
				BatchSize:     tc.batchSize,
			}))
			defer cleanup()

			tc.setup(t, s)

			query := &pb.HistoryQuery{
				PubsubTopic: pubSubTopic,
			}
			expectQueryMessagesEventually(t, s, query, tc.expected)
		})
	}
}

func expectQueryMessagesEventually(t *testing.T, s *XmtpStore, query *pb.HistoryQuery, expectedMsgs []*pb.WakuMessage) []*pb.WakuMessage {
	var msgs []*pb.WakuMessage
	require.Eventually(t, func() bool {
		msgs = queryMessages(t, s, query)
		return len(msgs) == len(expectedMsgs)
	}, 3*time.Second, 500*time.Millisecond, "expected %d == %d", len(msgs), len(expectedMsgs))
	require.ElementsMatch(t, expectedMsgs, msgs)
	return msgs
}

func queryMessages(t *testing.T, s *XmtpStore, query *pb.HistoryQuery) []*pb.WakuMessage {
	res, err := s.FindMessages(query)
	require.NoError(t, err)
	return res.Messages
}
