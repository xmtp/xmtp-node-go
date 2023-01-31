package crdt2

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
)

func Test_BasicBroadcast(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	net := newNetwork(t, ctx, 5, 1)
	net.Publish(0, t0, "hi")
	net.AssertEventuallyConsistent(time.Second)
}

// In-memory broadcaster that uses channels to broadcast Events between Nodes.
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

func (b *ChanBroadcaster) NewTopic(name string, n *Node) TopicBroadcaster {
	return NewTopicChanBroadcaster(b, name, n)
}

func (b *ChanBroadcaster) Broadcast(ev *Event, from *Node) {
	for sub := range b.subscribers {
		if sub == from {
			continue
		}
		t := sub.topics[ev.ContentTopic]
		if t == nil {
			t = sub.newTopic(ev.ContentTopic)
		}
		t.pendingReceiveEvents <- ev
	}
}

func (b *ChanBroadcaster) AddNode(n *Node) {
	b.subscribers[n] = true
}

func (b *ChanBroadcaster) RemoveNode(n *Node) {
	delete(b.subscribers, n)
}

type TopicChanBroadcaster struct {
	*ChanBroadcaster
	node *Node
	log  *zap.Logger
}

func NewTopicChanBroadcaster(b *ChanBroadcaster, name string, n *Node) *TopicChanBroadcaster {
	return &TopicChanBroadcaster{
		node:            n,
		log:             n.log.Named(name),
		ChanBroadcaster: b,
	}
}

func (tb *TopicChanBroadcaster) Broadcast(ev *Event) {
	tb.log.Debug("broadcasting", zapCid("event", ev.cid))
	tb.ChanBroadcaster.Broadcast(ev, tb.node)
}
