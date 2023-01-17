package crdt2

import mh "github.com/multiformats/go-multihash"

type TopicStore interface {
	NewEvent(payload []byte) (*Event, error)
	AddHead(*Event) (bool, error)
	RemoveHead(mh.Multihash) (bool, error)
}
