package crdt2

type Broadcaster interface {
	Broadcast(*Event)
	Events() <-chan *Event
}
