package crdt2

import (
	mh "github.com/multiformats/go-multihash"
	messagev1 "github.com/xmtp/proto/v3/go/message_api/v1"
	"go.uber.org/zap"
)

type NodeStore interface {
	// NewTopic creates a TopicStore for the specified topic.
	NewTopic(name string, log *zap.Logger) TopicStore
}

type TopicStore interface {
	// NewEvent creates and stores a new Event,
	// making the current heads its links and
	// replacing the heads with the new Event.
	// Return the new Event.
	NewEvent(*messagev1.Envelope) (*Event, error)
	// AddEvent stores the Event if it isn't know yet,
	// Return whether it was added.
	AddEvent(ev *Event) (added bool, err error)
	// AddHead stores the Event if it isn't know yet,
	// and add it to the heads
	// Return whether it was added.
	AddHead(ev *Event) (added bool, err error)
	// RemoveHead removes the CID from heads if it's there.
	// Return whether it was actually removed.
	RemoveHead(cid mh.Multihash) (removed bool, err error)
	// Get returns the Event based on its CID, nil if absent.
	// This is just for testing, not needed for the protocol implementation
	Get(cid mh.Multihash) (*Event, error)
}
