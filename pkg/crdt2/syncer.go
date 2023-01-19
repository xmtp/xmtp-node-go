package crdt2

import mh "github.com/multiformats/go-multihash"

type NodeSyncer interface {
	NewTopic(name string) TopicSyncer
}

type TopicSyncer interface {
	Fetch([]mh.Multihash) ([]*Event, error)
	// FetchAll() ([]*Event, error) // used to seed new nodes
}
