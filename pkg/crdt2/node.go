package crdt2

import (
	"context"
	"errors"

	mh "github.com/multiformats/go-multihash"
	messagev1 "github.com/xmtp/proto/v3/go/message_api/v1"
)

var TODO = errors.New("Not Yet Implemented")

type Node struct {
	ctx    context.Context
	Topics map[string]*Topic

	Store
	Syncer
	Broadcaster
}

func NewNode(ctx context.Context, store Store, syncer Syncer, bc Broadcaster) *Node {
	return &Node{
		ctx:         ctx,
		Topics:      make(map[string]*Topic),
		Store:       store,
		Syncer:      syncer,
		Broadcaster: bc,
	}
}

func (n *Node) NewTopic(name string) *Topic {
	t := NewTopic(name,
		n.Store.NewTopic(name),
		n.Syncer.NewTopic(name),
		n.Broadcaster.NewTopic(name))
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
