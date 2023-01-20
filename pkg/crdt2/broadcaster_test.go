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
	net := NewNetwork(t, ctx, 5, 1)
	net.PublishT0(0, "hi")
	net.AssertEventuallyConsistent(time.Second)
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

func (ps *ChanBroadcaster) RemoveNode(n *Node) {
	delete(ps.subscribers, n)
}

type TopicChanBroadcaster struct {
	*ChanBroadcaster
	name   string
	log    *zap.Logger
	events chan *Event
}

func NewTopicChanBroadcaster(ps *ChanBroadcaster, name string, log *zap.Logger) *TopicChanBroadcaster {
	return &TopicChanBroadcaster{
		log:             log,
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
