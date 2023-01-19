package crdt2

import "go.uber.org/zap"

type NodeBroadcaster interface {
	NewTopic(name string, log *zap.Logger) TopicBroadcaster
}

type TopicBroadcaster interface {
	Broadcast(*Event)
	Events() <-chan *Event
}
