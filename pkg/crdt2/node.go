package crdt2

import (
	"context"
	"errors"

	mh "github.com/multiformats/go-multihash"
	messagev1 "github.com/xmtp/proto/v3/go/message_api/v1"
	"go.uber.org/zap"
)

var TODO = errors.New("Not Yet Implemented")

type Node struct {
	ctx context.Context
	log *zap.Logger

	Topics map[string]*Topic

	NodeStore
	NodeSyncer
	NodeBroadcaster
}

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

func (n *Node) NewTopic(name string) *Topic {
	t := NewTopic(name,
		n.NodeStore.NewTopic(name),
		n.NodeSyncer.NewTopic(name),
		n.NodeBroadcaster.NewTopic(name))
	n.Topics[name] = t
	t.Start(n.ctx)
	return t
}

func (n *Node) Publish(ctx context.Context, env *messagev1.Envelope) (*Event, error) {
	return n.Topics[env.ContentTopic].Publish(ctx, env)
}

func (n *Node) Get(topic string, cid mh.Multihash) (*Event, error) {
	return n.Topics[topic].Get(cid)
}
