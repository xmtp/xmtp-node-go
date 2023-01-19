package crdt2

import (
	"context"

	mh "github.com/multiformats/go-multihash"
	messagev1 "github.com/xmtp/proto/v3/go/message_api/v1"
	"go.uber.org/zap"
)

type Topic struct {
	name              string
	pendingEvents     chan *Event
	pendingLinkEvents chan *Event
	pendingLinks      chan mh.Multihash
	log               *zap.Logger

	TopicStore
	TopicSyncer
	TopicBroadcaster
}

func NewTopic(name string, log *zap.Logger, store TopicStore, syncer TopicSyncer, bc TopicBroadcaster) *Topic {
	return &Topic{
		name:              name,
		pendingEvents:     make(chan *Event, 20),
		pendingLinkEvents: make(chan *Event, 20),
		pendingLinks:      make(chan mh.Multihash, 20),
		log:               log,
		TopicStore:        store,
		TopicSyncer:       syncer,
		TopicBroadcaster:  bc,
	}
}

func (t *Topic) Publish(ctx context.Context, env *messagev1.Envelope) (*Event, error) {
	ev, err := t.NewEvent(env)
	if err != nil {
		return nil, err
	}
	t.Broadcast(ev)
	return ev, nil
}

func (t *Topic) Query(ctx context.Context, req *messagev1.QueryRequest) ([]*messagev1.Envelope, *messagev1.PagingInfo, error) {
	return nil, nil, TODO
}

func (t *Topic) Start(ctx context.Context) {
	go t.receiveLoop(ctx)
	go t.syncLoop(ctx)
}

func (t *Topic) receiveLoop(ctx context.Context) {
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case ev := <-t.pendingEvents:
			t.AddHead(ev)
		case ev := <-t.Events():
			t.AddHead(ev)
		}
	}
}

func (t *Topic) syncLoop(ctx context.Context) {
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case ev := <-t.pendingLinkEvents:
			t.addEvent(ev)
		case cid := <-t.pendingLinks:
			t.log.Debug("checking link", zap.String("link", cid.String()))
			haveAlready, err := t.RemoveHead(cid)
			if err != nil {
				// requeue for later
				// TODO: may need a delay
				t.pendingLinks <- cid
				continue
			}
			if haveAlready {
				continue
			}
			evs, err := t.Fetch([]mh.Multihash{cid})
			if err != nil {
				// requeue for later
				// TODO: this will need refinement for invalid, missing cids etc.
				t.pendingLinks <- cid
			}
			for _, ev := range evs {
				t.addEvent(ev)
			}
		}
	}
}

func (t *Topic) addHead(ev *Event) {
	t.log.Debug("adding event", zap.String("event", ev.cid.String()))
	added, err := t.AddHead(ev)
	if err != nil {
		// requeue for later
		// TODO: may need a delay
		t.pendingEvents <- ev
	}
	if added {
		for _, link := range ev.links {
			t.pendingLinks <- link
		}
	}
}

func (t *Topic) addEvent(ev *Event) {
	t.log.Debug("adding link event", zap.String("event", ev.cid.String()))
	added, err := t.AddEvent(ev)
	if err != nil {
		// requeue for later
		// TODO: may need a delay
		t.pendingLinkEvents <- ev
	}
	if added {
		for _, link := range ev.links {
			t.pendingLinks <- link
		}
	}
}
