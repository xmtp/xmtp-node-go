package crdt2

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	messagev1 "github.com/xmtp/proto/v3/go/message_api/v1"
	"go.uber.org/zap"
)

func Test_BasicBroadcast(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	nodes := NewNetwork(t, ctx, 5, 1)
	ev, err := nodes[0].Publish(ctx, &messagev1.Envelope{TimestampNs: 1, ContentTopic: "t0", Message: []byte("hi")})
	assert.NoError(t, err)
	require.Eventually(t, func() bool {
		for i := 1; i < 3; i++ {
			ev, err := nodes[i].Get("t0", ev.cid)
			if err != nil || ev == nil {
				return false
			}
		}
		return true
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
