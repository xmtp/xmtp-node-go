package crdt2

import (
	"context"

	mh "github.com/multiformats/go-multihash"
	messagev1 "github.com/xmtp/proto/v3/go/message_api/v1"
)

type Topic struct {
	heads         *mh.Set
	pendingEvents chan *Event
	pendingCids   chan mh.Multihash

	Store
	Syncer
	Broadcaster
}

func (t *Topic) Publish(ctx context.Context, env *messagev1.Envelope) error {
	ev, err := NewEvent(env.Message, t.heads.All())
	if err != nil {
		return err
	}
	// TODO: this needs to be atomic
	if err = t.Put(ev); err != nil {
		return err
	}
	t.heads = mh.NewSet()
	t.heads.Add(ev.cid)
	// ... to here
	t.Broadcast(ev)
	return nil
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
		case ev := <-t.Events():
			t.receiveEvent(ev)
		}
	}
}

func (t *Topic) syncLoop(ctx context.Context) {
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case ev := <-t.pendingEvents:
			t.receiveEvent(ev)
		case cid := <-t.pendingCids:
			have, err := t.Has(cid)
			if err != nil {
				t.pendingCids <- cid
				continue
			}
			if have {
				t.heads.Remove(cid)
				continue
			}
			evs, err := t.Fetch([]mh.Multihash{cid})
			if err != nil {
				// requeue for later
				// TODO: this will need refinement for invalid, missing cids etc.
				t.pendingCids <- cid
			}
			for _, ev := range evs {
				t.receiveEvent(ev)
			}
		}
	}
}

func (t *Topic) receiveEvent(ev *Event) {
	have, err := t.Has(ev.cid)
	if err != nil {
		// requeue the event for later
		// this loop might be too tight if store is in trouble
		t.pendingEvents <- ev
	}
	if have {
		return
	}

	// TODO: this needs to happen atomically
	if err := t.Put(ev); err != nil {
		// requeue for later
		t.pendingEvents <- ev
	}
	t.heads.Add(ev.cid)
	// ... up to here

	for _, link := range ev.links {
		t.pendingCids <- link
	}
}
