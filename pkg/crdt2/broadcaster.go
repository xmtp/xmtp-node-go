package crdt2

type NodeBroadcaster interface {
	NewTopic(name string) TopicBroadcaster
}

type TopicBroadcaster interface {
	Broadcast(*Event)
	Events() chan *Event
}
