package crdt2

import (
	"context"

	mh "github.com/multiformats/go-multihash"
	messagev1 "github.com/xmtp/proto/v3/go/message_api/v1"
	"go.uber.org/zap"
)

// Topic manages the DAG of a topic replica.
// It implements the topic API, as well as the
// replication mechanism using the store, broadcaster and syncer.
type Topic struct {
	name                 string            // the topic name
	pendingReceiveEvents chan *Event       // broadcasted events that were received from the network but not processed yet
	pendingSyncEvents    chan *Event       // missing events that were fetched from the network but not processed yet
	pendingLinks         chan mh.Multihash // missing links that were discovered but not successfully fetched yet
	log                  *zap.Logger

	TopicStore
	TopicSyncer
	TopicBroadcaster
}

// Creates a new topic replica
func NewTopic(name string, log *zap.Logger, store TopicStore, syncer TopicSyncer, bc TopicBroadcaster) *Topic {
	return &Topic{
		name: name,
		// TODO: tuning the channel sizes will likely be important
		// current implementation can lock up if the channels fill up.
		pendingReceiveEvents: make(chan *Event, 20),
		pendingSyncEvents:    make(chan *Event, 20),
		pendingLinks:         make(chan mh.Multihash, 20),
		log:                  log,
		TopicStore:           store,
		TopicSyncer:          syncer,
		TopicBroadcaster:     bc,
	}
}

// Publish adopts a new message into a topic and broadcasts it to the network.
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

// Start the replication mechanisms of the topic.
func (t *Topic) Start(ctx context.Context) {
	go t.receiveLoop(ctx)
	go t.syncLoop(ctx)
}

// receiveLoop processes incoming Event broadcasts.
func (t *Topic) receiveLoop(ctx context.Context) {
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case ev := <-t.pendingReceiveEvents:
			t.addHead(ev)
		case ev := <-t.Events():
			t.addHead(ev)
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
		t.pendingReceiveEvents <- ev
	}
	if added {
		for _, link := range ev.links {
			t.pendingLinks <- link
		}
	}
}

// syncLoop implements topic syncing
func (t *Topic) syncLoop(ctx context.Context) {
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case ev := <-t.pendingSyncEvents:
			t.addEvent(ev)
		case cid := <-t.pendingLinks:
			// t.log.Debug("checking link", zapCid("link", cid))
			// If the CID is in heads, it should be removed because
			// we have an event that points to it.
			// We also don't need to fetch it since we already have it.
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
					// requeue missing links
					t.pendingLinks <- cids[i]
					continue
				}
				t.addEvent(ev)
			}
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
		t.pendingSyncEvents <- ev
	}
	if added {
		for _, link := range ev.links {
			// TODO: if the channel is full, this will lock up the loop
			t.pendingLinks <- link
		}
	}
}
