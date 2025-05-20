package subscriptions

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	proto "github.com/xmtp/xmtp-node-go/pkg/proto/message_api/v1"
	testutils "github.com/xmtp/xmtp-node-go/pkg/testing"
)

const (
	TOPIC_1 = "/xmtp/0/topic1"
	TOPIC_2 = "/xmtp/0/topic2"
)

func buildEnvelope(topic string) *proto.Envelope {
	return &proto.Envelope{
		ContentTopic: topic,
	}
}

func requireMessage(t *testing.T, sub *Subscription) *proto.Envelope {
	select {
	case msg := <-sub.MessagesCh:
		return msg
	case <-time.After(100 * time.Millisecond):
		require.Fail(t, "expected message on sub")
		return nil
	}
}

func TestSimpleDispatch(t *testing.T) {
	dispatcher := NewSubscriptionDispatcher(testutils.NewLog(t))

	sub1 := dispatcher.Subscribe(map[string]bool{TOPIC_1: true})

	dispatcher.HandleEnvelope(buildEnvelope(TOPIC_1))

	msg := requireMessage(t, sub1)
	require.Equal(t, msg.ContentTopic, TOPIC_1)

	dispatcher.HandleEnvelope(buildEnvelope(TOPIC_2))
	require.Len(t, sub1.MessagesCh, 0)
}

func TestWildcardDispatch(t *testing.T) {
	dispatcher := NewSubscriptionDispatcher(testutils.NewLog(t))

	sub := dispatcher.Subscribe(map[string]bool{WILDCARD_TOPIC: true})

	dispatcher.HandleEnvelope(buildEnvelope(TOPIC_1))

	msg := requireMessage(t, sub)
	require.Equal(t, msg.ContentTopic, TOPIC_1)

	// Ensure the dispatcher ignores invalid topics
	dispatcher.HandleEnvelope(buildEnvelope("invalid"))
	require.Len(t, sub.MessagesCh, 0)
}

func TestMultipleSubscriptions(t *testing.T) {
	dispatcher := NewSubscriptionDispatcher(testutils.NewLog(t))

	sub1 := dispatcher.Subscribe(map[string]bool{TOPIC_1: true})
	sub2 := dispatcher.Subscribe(map[string]bool{TOPIC_1: true})

	dispatcher.HandleEnvelope(buildEnvelope(TOPIC_1))

	msg1 := requireMessage(t, sub1)
	msg2 := requireMessage(t, sub2)
	require.Equal(t, msg1.ContentTopic, TOPIC_1)
	require.Equal(t, msg2.ContentTopic, TOPIC_1)
}
