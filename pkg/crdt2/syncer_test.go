package crdt2

import mh "github.com/multiformats/go-multihash"

type nilSyncer struct{}

func (_ *nilSyncer) NewTopic(name string) TopicSyncer {
	return &nilTopicSyncer{}
}

type nilTopicSyncer struct{}

func (*nilTopicSyncer) Fetch([]mh.Multihash) ([]*Event, error) {
	return nil, TODO
}
