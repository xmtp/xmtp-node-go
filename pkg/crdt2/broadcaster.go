package crdt2

type TopicBroadcaster interface {
	Broadcast(*Event)
	Events() <-chan *Event
}
