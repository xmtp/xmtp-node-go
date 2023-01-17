package crdt2

import (
	"context"

	mh "github.com/multiformats/go-multihash"
	messagev1 "github.com/xmtp/proto/v3/go/message_api/v1"
)

type Topic struct {
	pendingEvents chan *Event
	pendingCids   chan mh.Multihash

	TopicStore
	TopicSyncer
	TopicBroadcaster
}

func (t *Topic) Publish(ctx context.Context, env *messagev1.Envelope) error {
	ev, err := t.NewEvent(env.Message)
	if err != nil {
		return err
	}
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
			haveAlready, err := t.RemoveHead(cid)
			if err != nil {
				// retry later
				t.pendingCids <- cid
				continue
			}
			if haveAlready {
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
	added, err := t.AddHead(ev)
	if err != nil {
		// requeue for later
		t.pendingEvents <- ev
	}

	if added {
		for _, link := range ev.links {
			t.pendingCids <- link
		}
	}
}
