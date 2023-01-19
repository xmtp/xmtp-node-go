package store

import (
	"testing"
	"time"

	"github.com/status-im/go-waku/waku/v2/protocol/pb"
	test "github.com/xmtp/xmtp-node-go/pkg/testing"
)

func TestStore_Cleaner_DeletesNonXMTPMessages(t *testing.T) {
	t.Parallel()

	s, cleanup := newTestStore(t, WithCleaner(CleanerOptions{
		Enable:        true,
		Period:        1 * time.Second,
		RetentionDays: 3,
	}))
	defer cleanup()

	c := newTestClient(t, s.host.ID())
	addStoreProtocol(t, c.host, s.host)

	pubSubTopic := "test-" + test.RandomStringLower(5)

	fourDaysAgo := time.Now().UTC().Add(-4 * 24 * time.Hour).UnixNano()
	storeMessageWithTime(t, s, test.NewMessage("topic1", 1, "msg1"), pubSubTopic, fourDaysAgo)
	storeMessageWithTime(t, s, test.NewMessage("/xmtp/topic2", 2, "msg2"), pubSubTopic, fourDaysAgo)
	storeMessageWithTime(t, s, test.NewMessage("topic3", 3, "msg4"), pubSubTopic, fourDaysAgo)
	storeMessageWithTime(t, s, test.NewMessage("/xmtp/topic4", 4, "msg4"), pubSubTopic, fourDaysAgo)

	storeMessage(t, s, test.NewMessage("topic5", 5, "msg5"), pubSubTopic)
	storeMessage(t, s, test.NewMessage("/xmtp/topic6", 6, "msg6"), pubSubTopic)
	storeMessage(t, s, test.NewMessage("topic7", 7, "msg7"), pubSubTopic)
	storeMessage(t, s, test.NewMessage("/xmtp/topic8", 8, "msg8"), pubSubTopic)

	query := &pb.HistoryQuery{
		PubsubTopic: pubSubTopic,
	}
	expectQueryMessagesEventually(t, c, query, []*pb.WakuMessage{
		test.NewMessage("/xmtp/topic2", 2, "msg2"),
		test.NewMessage("/xmtp/topic4", 4, "msg4"),
		test.NewMessage("topic5", 5, "msg5"),
		test.NewMessage("/xmtp/topic6", 6, "msg6"),
		test.NewMessage("topic7", 7, "msg7"),
		test.NewMessage("/xmtp/topic8", 8, "msg8"),
	})
}
