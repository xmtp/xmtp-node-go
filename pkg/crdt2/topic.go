package crdt2

import (
	"context"

	mh "github.com/multiformats/go-multihash"
	messagev1 "github.com/xmtp/proto/v3/go/message_api/v1"
)

type Topic struct {
	heads []mh.Multihash

	Store
	Syncer
	Broadcaster
}

func (t *Topic) Publish(ctx context.Context, env *messagev1.Envelope) error {
	ev, err := NewEvent(env.Message, t.heads)
	if err != nil {
		return err
	}
	if err = t.Put(ev); err != nil {
		return err
	}
	t.heads = []mh.Multihash{ev.cid}
	t.Broadcast(ev)
	return nil
}

func (t *Topic) Query(ctx context.Context, req *messagev1.QueryRequest) ([]*messagev1.Envelope, *messagev1.PagingInfo, error) {
	return nil, nil, TODO
}

func (t *Topic) Start(ctx context.Context) {
	go t.eventLoop(ctx)
}

func (t *Topic) eventLoop(ctx context.Context) {
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

func (t *Topic) receiveEvent(ev *Event) {

}
