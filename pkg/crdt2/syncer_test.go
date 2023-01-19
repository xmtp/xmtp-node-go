package crdt2

import (
	mh "github.com/multiformats/go-multihash"
	"go.uber.org/zap"
)

type nilSyncer struct{}

func (_ *nilSyncer) NewTopic(name string, log *zap.Logger) TopicSyncer {
	return &nilTopicSyncer{}
}

type nilTopicSyncer struct{}

func (*nilTopicSyncer) Fetch([]mh.Multihash) ([]*Event, error) {
	return nil, TODO
}
