package store

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	messagev1 "github.com/xmtp/xmtp-node-go/pkg/proto/message_api/v1"
	test "github.com/xmtp/xmtp-node-go/pkg/testing"
)

func TestStore_Cleaner(t *testing.T) {
	tcs := []struct {
		name      string
		batchSize int
		setup     func(t *testing.T, s *Store)
		expected  []*messagev1.Envelope
	}{
		{
			name:      "no messages",
			batchSize: 5,
			setup:     func(t *testing.T, s *Store) {},
			expected:  []*messagev1.Envelope{},
		},
		{
			name:      "less than batch size",
			batchSize: 5,
			setup: func(t *testing.T, s *Store) {
				fourDaysAgo := time.Now().UTC().Add(-4 * 24 * time.Hour).UnixNano()
				storeMessageWithTime(t, s, test.NewEnvelope("topic1", 1, "msg1"), fourDaysAgo)
				storeMessageWithTime(t, s, test.NewEnvelope("/xmtp/topic2", 2, "msg2"), fourDaysAgo)
				storeMessageWithTime(t, s, test.NewEnvelope("topic3", 3, "msg3"), fourDaysAgo)
				storeMessageWithTime(t, s, test.NewEnvelope("/xmtp/topic4", 4, "msg4"), fourDaysAgo)

				storeMessage(t, s, test.NewEnvelope("topic5", 5, "msg5"))
				storeMessage(t, s, test.NewEnvelope("/xmtp/topic6", 6, "msg6"))
				storeMessage(t, s, test.NewEnvelope("topic7", 7, "msg7"))
				storeMessage(t, s, test.NewEnvelope("/xmtp/topic8", 8, "msg8"))
			},
			expected: []*messagev1.Envelope{
				test.NewEnvelope("topic1", 1, "msg1"),
				test.NewEnvelope("/xmtp/topic2", 2, "msg2"),
				test.NewEnvelope("topic3", 3, "msg3"),
				test.NewEnvelope("/xmtp/topic4", 4, "msg4"),
				test.NewEnvelope("topic5", 5, "msg5"),
				test.NewEnvelope("/xmtp/topic6", 6, "msg6"),
				test.NewEnvelope("topic7", 7, "msg7"),
				test.NewEnvelope("/xmtp/topic8", 8, "msg8"),
			},
		},
		{
			name:      "equal to batch size",
			batchSize: 3,
			setup: func(t *testing.T, s *Store) {
				fourDaysAgo := time.Now().UTC().Add(-4 * 24 * time.Hour).UnixNano()
				storeMessageWithTime(t, s, test.NewEnvelope("topic1", 1, "msg1"), fourDaysAgo)
				storeMessageWithTime(t, s, test.NewEnvelope("/xmtp/topic2", 2, "msg2"), fourDaysAgo)
				storeMessageWithTime(t, s, test.NewEnvelope("topic3", 3, "msg3"), fourDaysAgo)
				storeMessageWithTime(t, s, test.NewEnvelope("/xmtp/topic4", 4, "msg4"), fourDaysAgo)
				storeMessageWithTime(t, s, test.NewEnvelope("topic5", 5, "msg5"), fourDaysAgo)
				storeMessageWithTime(t, s, test.NewEnvelope("/xmtp/topic6", 6, "msg6"), fourDaysAgo)

				storeMessage(t, s, test.NewEnvelope("topic7", 7, "msg7"))
				storeMessage(t, s, test.NewEnvelope("/xmtp/topic8", 8, "msg8"))
			},
			expected: []*messagev1.Envelope{
				test.NewEnvelope("/xmtp/topic2", 2, "msg2"),
				test.NewEnvelope("/xmtp/topic4", 4, "msg4"),
				test.NewEnvelope("/xmtp/topic6", 6, "msg6"),
				test.NewEnvelope("topic7", 7, "msg7"),
				test.NewEnvelope("/xmtp/topic8", 8, "msg8"),
			},
		},
		{
			name:      "greater than batch size",
			batchSize: 2,
			setup: func(t *testing.T, s *Store) {
				fourDaysAgo := time.Now().UTC().Add(-4 * 24 * time.Hour).UnixNano()
				storeMessageWithTime(t, s, test.NewEnvelope("topic1", 1, "msg1"), fourDaysAgo)
				storeMessageWithTime(t, s, test.NewEnvelope("/xmtp/topic2", 2, "msg2"), fourDaysAgo)
				storeMessageWithTime(t, s, test.NewEnvelope("topic3", 3, "msg3"), fourDaysAgo)
				storeMessageWithTime(t, s, test.NewEnvelope("/xmtp/topic4", 4, "msg4"), fourDaysAgo)
				storeMessageWithTime(t, s, test.NewEnvelope("topic5", 5, "msg5"), fourDaysAgo)
				storeMessageWithTime(t, s, test.NewEnvelope("/xmtp/topic6", 6, "msg6"), fourDaysAgo)

				storeMessage(t, s, test.NewEnvelope("topic7", 7, "msg7"))
				storeMessage(t, s, test.NewEnvelope("/xmtp/topic8", 8, "msg8"))
			},
			expected: []*messagev1.Envelope{
				test.NewEnvelope("/xmtp/topic2", 2, "msg2"),
				test.NewEnvelope("/xmtp/topic4", 4, "msg4"),
				test.NewEnvelope("topic5", 5, "msg5"),
				test.NewEnvelope("/xmtp/topic6", 6, "msg6"),
				test.NewEnvelope("topic7", 7, "msg7"),
				test.NewEnvelope("/xmtp/topic8", 8, "msg8"),
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

			query := &messagev1.QueryRequest{}
			expectQueryMessagesEventually(t, s, query, tc.expected)
		})
	}
}

func expectQueryMessagesEventually(t *testing.T, s *Store, query *messagev1.QueryRequest, expectedMsgs []*messagev1.Envelope) []*messagev1.Envelope {
	t.Helper()
	var msgs []*messagev1.Envelope
	require.Eventually(t, func() bool {
		msgs = queryMessages(t, s, query)
		return len(msgs) == len(expectedMsgs)
	}, 3*time.Second, 500*time.Millisecond, "expected %d", len(expectedMsgs))
	require.ElementsMatch(t, expectedMsgs, msgs)
	return msgs
}

func queryMessages(t *testing.T, s *Store, query *messagev1.QueryRequest) []*messagev1.Envelope {
	t.Helper()
	res, err := s.Query(query)
	require.NoError(t, err)
	return res.Envelopes
}
