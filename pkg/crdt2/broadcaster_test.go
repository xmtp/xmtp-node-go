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

func Test_BasicBroadcast(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	log := test.NewLog(t)
	ps := NewChanBroadcaster(log)
	n1 := NewNode(ctx, log.Named("n1"), NewMapStore(), (*nilSyncer)(nil), ps)
	ps.AddNode(n1)
	n2 := NewNode(ctx, log.Named("n2"), NewMapStore(), (*nilSyncer)(nil), ps)
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

// In-memory broadcaster
type ChanBroadcaster struct {
	log         *zap.Logger
	subscribers map[*Node]bool
}

func NewChanBroadcaster(log *zap.Logger) *ChanBroadcaster {
	return &ChanBroadcaster{
		log:         log.Named("chanbc"),
		subscribers: make(map[*Node]bool),
	}
}

func (ps *ChanBroadcaster) NewTopic(name string, log *zap.Logger) TopicBroadcaster {
	return NewTopicChanBroadcaster(ps, name, log)
}

func (ps *ChanBroadcaster) Broadcast(ev *Event) {
	for sub := range ps.subscribers {
		if t := sub.Topics[ev.ContentTopic]; t != nil {
			t.TopicBroadcaster.(*TopicChanBroadcaster).events <- ev
		}
	}
}

func (ps *ChanBroadcaster) AddNode(n *Node) {
	ps.subscribers[n] = true
}

type TopicChanBroadcaster struct {
	*ChanBroadcaster
	name   string
	log    *zap.Logger
	events chan *Event
}

func NewTopicChanBroadcaster(ps *ChanBroadcaster, name string, log *zap.Logger) *TopicChanBroadcaster {
	return &TopicChanBroadcaster{
		log:             log.Named(name),
		name:            name,
		ChanBroadcaster: ps,
		events:          make(chan *Event, 20),
	}
}

func (tb *TopicChanBroadcaster) Broadcast(ev *Event) {
	tb.log.Debug("broadcasting", zap.Stringer("event", ev.cid))
	tb.ChanBroadcaster.Broadcast(ev)
}

func (tb *TopicChanBroadcaster) Events() <-chan *Event {
	return tb.events
}
