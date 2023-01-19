package crdt2

import (
	mh "github.com/multiformats/go-multihash"
	messagev1 "github.com/xmtp/proto/v3/go/message_api/v1"
)

type NodeStore interface {
	NewTopic(name string) TopicStore
}

type TopicStore interface {
	NewEvent(*messagev1.Envelope) (*Event, error)
	AddHead(*Event) (bool, error)
	RemoveHead(mh.Multihash) (bool, error)
	// Get returns the Event based on its CID, nil if absent.
	// This is just for testing, not needed for the protocol implementation
	Get(cid mh.Multihash) (*Event, error)
}
