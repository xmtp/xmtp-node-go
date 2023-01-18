package crdt2

type Broadcaster interface {
	NewTopic(name string) TopicBroadcaster
}

type TopicBroadcaster interface {
	Broadcast(*Event)
	Events() chan *Event
}
