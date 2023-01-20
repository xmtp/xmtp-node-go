package crdt2

import (
	"context"
	"math/rand"
	"testing"
	"time"

	mh "github.com/multiformats/go-multihash"
	"go.uber.org/zap"
)

func Test_BasicSyncing(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	net := NewNetwork(t, ctx, 3, 1)
	net.PublishT0(0, "hi")
	net.PublishT0(1, "hi back")
	net.AssertEventuallyConsistent(time.Second)
	net.WithSuspendedTopic(1, "t0", func(n *Node) {
		net.PublishT0(2, "oh hello")
		net.PublishT0(2, "how goes")
		net.PublishT0(1, "how are you")
	})
	net.AssertEventuallyConsistent(time.Second, 1)
	net.PublishT0(0, "not bad")
	net.AssertEventuallyConsistent(time.Second)
}

// In-memory syncer
type randomSyncer struct {
	nodes []*Node
}

func NewRandomSyncer() *randomSyncer {
	return &randomSyncer{}
}

func (s *randomSyncer) AddNode(n *Node) {
	s.nodes = append(s.nodes, n)
}

func (s *randomSyncer) NewTopic(name string, log *zap.Logger) TopicSyncer {
	return &randomTopicSyncer{
		randomSyncer: s,
		topic:        name,
		log:          log,
	}
}

type randomTopicSyncer struct {
	*randomSyncer
	topic string
	log   *zap.Logger
}

func (s *randomTopicSyncer) Fetch(cids []mh.Multihash) (results []*Event, err error) {
	node := s.nodes[rand.Intn(len(s.nodes))]
	s.log.Debug("fetching", zapCids(cids...))
	for _, cid := range cids {
		ev, err := node.Get(s.topic, cid)
		if err != nil {
			return nil, err
		}
		results = append(results, ev)
	}
	return results, nil
}
