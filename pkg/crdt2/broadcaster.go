package crdt2

import "go.uber.org/zap"

// NodeBroadcaster manages the overall broadcasting capacity of a Node
type NodeBroadcaster interface {
	// NewTopic creates a broadcaster for a specific topic
	NewTopic(name string, log *zap.Logger) TopicBroadcaster
}

// TopicBroadcaster manages broadcasts for a given topic
type TopicBroadcaster interface {
	// Broadcast sends an Event out to the network
	Broadcast(*Event)
	// Events channel is used to receive Events broadcasted by other nodes
	Events() <-chan *Event
}
