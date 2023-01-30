package crdt2

import (
	"context"
	"errors"
	"sync"

	mh "github.com/multiformats/go-multihash"
	messagev1 "github.com/xmtp/proto/v3/go/message_api/v1"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

var TODO = errors.New("Not Yet Implemented")
var UnknownTopic = errors.New("Unknown Topic")

// Node represents a peer in the XMTP network.
// Node hosts a set of Topics and provides the required
// supporting facilities (store, syncer, broadcaster).
type Node struct {
	ctx    context.Context
	Cancel context.CancelFunc
	log    *zap.Logger

	sync.RWMutex // to protect the topics map
	topics       map[string]*Topic

	NodeStore
	NodeSyncer
	NodeBroadcaster
}

// NewNode creates a new network node.
func NewNode(ctx context.Context, log *zap.Logger, store NodeStore, syncer NodeSyncer, bc NodeBroadcaster) (*Node, error) {
	ctx, cancel := context.WithCancel(ctx)
	node := &Node{
		ctx:             ctx,
		Cancel:          cancel,
		log:             log,
		topics:          make(map[string]*Topic),
		NodeStore:       store,
		NodeSyncer:      syncer,
		NodeBroadcaster: bc,
	}
	topics, err := store.Topics()
	if err != nil {
		return nil, err
	}
	grp, ctx := errgroup.WithContext(ctx)
	grp.SetLimit(1000) // up to 1000 topic bootstraps in parallel
	for _, n := range topics {
		topic := n
		grp.Go(func() (err error) {
			var t *Topic
			{
				node.Lock()
				defer node.Unlock()
				t = node.newTopic(topic)
			}
			return t.bootstrap(ctx)
		})
	}
	if err := grp.Wait(); err != nil {
		cancel()
		return nil, err
	}
	return node, nil
}

// Publish sends a new message out to the network.
func (n *Node) Publish(ctx context.Context, env *messagev1.Envelope) (*Event, error) {
	topic := n.getOrCreateTopic(env.ContentTopic)
	return topic.Publish(ctx, env)
}

// Get retrieves an Event for given Topic.
func (n *Node) Get(topic string, cid mh.Multihash) (*Event, error) {
	var t *Topic
	{
		n.RLock()
		defer n.RUnlock()
		t = n.topics[topic]
		if t == nil {
			return nil, UnknownTopic
		}
	}
	return t.Get(cid)
}

// Count returns count of all events on the Node.
func (n *Node) Count() (count int, err error) {
	n.RLock()
	defer n.RUnlock()
	for _, t := range n.topics {
		tc, err := t.Count()
		if err != nil {
			return 0, err
		}
		count += tc
	}
	return count, nil
}

// getOrCreateTopic MUST NOT be called before topic bootstrap is complete
// to avoid creating empty topics that weren't bootstrapped.
func (n *Node) getOrCreateTopic(topic string) *Topic {
	n.Lock()
	defer n.Unlock()
	t := n.topics[topic]
	if t == nil {
		t = n.newTopic(topic)
	}
	return t
}

// newTopic adds a topic to the Node.
// MUST be called with a write lock!
func (n *Node) newTopic(name string) *Topic {
	t := NewTopic(
		n.ctx,
		name,
		n.log.Named(name),
		n.NodeStore.NewTopic(name, n),
		n.NodeSyncer.NewTopic(name, n),
		n.NodeBroadcaster.NewTopic(name, n),
	)
	n.topics[name] = t
	return t
}
