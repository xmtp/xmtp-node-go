package crdt2

import (
	"context"
	"errors"

	mh "github.com/multiformats/go-multihash"
	messagev1 "github.com/xmtp/proto/v3/go/message_api/v1"
	"go.uber.org/zap"
)

var TODO = errors.New("Not Yet Implemented")
var UnknownTopic = errors.New("Unknown Topic")

// Node represents a node in the XMTP network.
// Node hosts a set of Topics and provides the required
// supporting facilities (store, syncer, broadcaster).
type Node struct {
	ctx context.Context
	log *zap.Logger

	Topics map[string]*Topic

	NodeStore
	NodeSyncer
	NodeBroadcaster
}

// NewNode creates a new network node.
func NewNode(ctx context.Context, log *zap.Logger, store NodeStore, syncer NodeSyncer, bc NodeBroadcaster) *Node {
	return &Node{
		ctx:             ctx,
		log:             log,
		Topics:          make(map[string]*Topic),
		NodeStore:       store,
		NodeSyncer:      syncer,
		NodeBroadcaster: bc,
	}
}

// NewTopic adds a topic to the Node.
func (n *Node) NewTopic(name string) *Topic {
	t := NewTopic(name,
		n.log.Named(name),
		n.NodeStore.NewTopic(name, n),
		n.NodeSyncer.NewTopic(name, n),
		n.NodeBroadcaster.NewTopic(name, n),
	)
	n.Topics[name] = t
	t.Start(n.ctx)
	return t
}

// Publish sends a new message out to the network.
func (n *Node) Publish(ctx context.Context, env *messagev1.Envelope) (*Event, error) {
	topic := n.Topics[env.ContentTopic]
	if topic == nil {
		return nil, UnknownTopic
	}
	return topic.Publish(ctx, env)
}

// Get retrieves an Event for given Topic.
func (n *Node) Get(topic string, cid mh.Multihash) (*Event, error) {
	t := n.Topics[topic]
	if t == nil {
		return nil, UnknownTopic
	}
	return t.Get(cid)
}

// Count returns count of all events on the Node.
func (n *Node) Count() (count int, err error) {
	for _, t := range n.Topics {
		tc, err := t.Count()
		if err != nil {
			return 0, err
		}
		count += tc
	}
	return count, nil
}
