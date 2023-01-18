package crdt2

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	messagev1 "github.com/xmtp/proto/v3/go/message_api/v1"
)

func Test_BasicPubSub(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ps := NewPubSub()
	n1 := NewNode(ctx, NewMapStore(), (*nilSyncer)(nil), ps)
	ps.AddNode(n1)
	n2 := NewNode(ctx, NewMapStore(), (*nilSyncer)(nil), ps)
	ps.AddNode(n2)
	n1.NewTopic("t1")
	n2.NewTopic("t1")
	ev, err := n1.Publish(ctx, &messagev1.Envelope{TimestampNs: 1, ContentTopic: "t1", Message: []byte("hi")})
	assert.NoError(t, err)
	require.Eventually(t, func() bool {
		ev, err := n2.Get("t1", ev.cid)
		if err != nil {
			return false
		}
		return ev != nil
	}, 100*time.Millisecond, 10*time.Millisecond)
}

// In-memory topic broadcaster
type PubSub struct {
	subscribers map[*Node]bool
}

func NewPubSub() *PubSub {
	return &PubSub{subscribers: make(map[*Node]bool)}
}

func (ps *PubSub) NewTopic(name string) TopicBroadcaster {
	return NewTopicPubSub(ps)
}

func (ps *PubSub) Broadcast(ev *Event) {
	for sub := range ps.subscribers {
		sub.Topics[ev.ContentTopic].Events() <- ev
	}
}

func (ps *PubSub) AddNode(n *Node) {
	ps.subscribers[n] = true
}

type TopicPubSub struct {
	*PubSub
	events chan *Event
}

func NewTopicPubSub(ps *PubSub) *TopicPubSub {
	return &TopicPubSub{
		PubSub: ps,
		events: make(chan *Event, 20),
	}
}

func (tb *TopicPubSub) Broadcast(ev *Event) {
	tb.PubSub.Broadcast(ev)
}

func (tb *TopicPubSub) Events() chan *Event {
	return tb.events
}
