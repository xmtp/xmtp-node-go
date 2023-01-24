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
			t.addHead(ev)
		case ev := <-t.Events():
			t.addHead(ev)
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
			// t.log.Debug("checking link", zapCid("link", cid))
			haveAlready, err := t.RemoveHead(cid)
			if err != nil {
				// requeue for later
				// TODO: may need a delay
				// TODO: if the channel is full, this will lock up the loop
				t.pendingLinks <- cid
				continue
			}
			if haveAlready {
				continue
			}
			t.log.Debug("fetching link", zapCid("link", cid))
			cids := []mh.Multihash{cid}
			evs, err := t.Fetch(cids)
			if err != nil {
				// requeue for later
				// TODO: this will need refinement for invalid, missing cids etc.
				// TODO: if the channel is full, this will lock up the loop
				t.pendingLinks <- cid
			}
			for i, ev := range evs {
				if ev == nil {
					t.pendingLinks <- cids[i]
					continue
				}
				t.addEvent(ev)
			}
		}
	}
}

func (t *Topic) addHead(ev *Event) {
	// t.log.Debug("adding event", zapCid("event", ev.cid))
	added, err := t.AddHead(ev)
	if err != nil {
		// requeue for later
		// TODO: may need a delay
		// TODO: if the channel is full, this will lock up the loop
		t.pendingEvents <- ev
	}
	if added {
		for _, link := range ev.links {
			t.pendingLinks <- link
		}
	}
}

func (t *Topic) addEvent(ev *Event) {
	// t.log.Debug("adding link event", zapCid("event", ev.cid))
	added, err := t.AddEvent(ev)
	if err != nil {
		// requeue for later
		// TODO: may need a delay
		// TODO: if the channel is full, this will lock up the loop
		t.pendingLinkEvents <- ev
	}
	if added {
		for _, link := range ev.links {
			// TODO: if the channel is full, this will lock up the loop
			t.pendingLinks <- link
		}
	}
}
