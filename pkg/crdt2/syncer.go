package crdt2

import (
	mh "github.com/multiformats/go-multihash"
	"go.uber.org/zap"
)

type NodeSyncer interface {
	NewTopic(name string, log *zap.Logger) TopicSyncer
}

type TopicSyncer interface {
	Fetch([]mh.Multihash) ([]*Event, error)
	// FetchAll() ([]*Event, error) // used to seed new nodes
}
