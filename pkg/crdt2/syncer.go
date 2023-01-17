package crdt2

import mh "github.com/multiformats/go-multihash"

type TopicSyncer interface {
	Fetch([]mh.Multihash) ([]*Event, error)
	// FetchAll() ([]*Event, error) // used to seed new nodes
}
