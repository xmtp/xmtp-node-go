package crdt2

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	messagev1 "github.com/xmtp/proto/v3/go/message_api/v1"
	test "github.com/xmtp/xmtp-node-go/pkg/testing"
	"go.uber.org/zap"
)

func Test_BasicPubSub(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	log := test.NewLog(t)
	ps := NewPubSub(log)
	n1 := NewNode(ctx, log.Named("n1"), NewMapStore(log.Named("n1")), (*nilSyncer)(nil), ps)
	ps.AddNode(n1)
	n2 := NewNode(ctx, log.Named("n2"), NewMapStore(log.Named("n2")), (*nilSyncer)(nil), ps)
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
	log         *zap.Logger
	subscribers map[*Node]bool
}

func NewPubSub(log *zap.Logger) *PubSub {
	return &PubSub{
		log:         log.Named("pubsub"),
		subscribers: make(map[*Node]bool),
	}
}

func (ps *PubSub) NewTopic(name string) TopicBroadcaster {
	return NewTopicPubSub(ps, name)
}

func (ps *PubSub) Broadcast(ev *Event) {
	for sub := range ps.subscribers {
		if t := sub.Topics[ev.ContentTopic]; t != nil {
			t.Events() <- ev
		}
	}
}

func (ps *PubSub) AddNode(n *Node) {
	ps.subscribers[n] = true
}

type TopicPubSub struct {
	*PubSub
	name   string
	log    *zap.Logger
	events chan *Event
}

func NewTopicPubSub(ps *PubSub, name string) *TopicPubSub {
	return &TopicPubSub{
		log:    ps.log.Named(name),
		name:   name,
		PubSub: ps,
		events: make(chan *Event, 20),
	}
}

func (tb *TopicPubSub) Broadcast(ev *Event) {
	tb.log.Debug("broadcasting", zap.Stringer("event", ev.cid))
	tb.PubSub.Broadcast(ev)
}

func (tb *TopicPubSub) Events() chan *Event {
	return tb.events
}
