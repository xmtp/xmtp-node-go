package crdt2

import (
	mh "github.com/multiformats/go-multihash"
	messagev1 "github.com/xmtp/proto/v3/go/message_api/v1"
)

type Store interface {
	NewTopic(name string) TopicStore
}

type TopicStore interface {
	NewEvent(*messagev1.Envelope) (*Event, error)
	AddHead(*Event) (bool, error)
	RemoveHead(mh.Multihash) (bool, error)
	Get(mh.Multihash) (*Event, error)
}
