package crdt2

type PubSub struct {
	subscribers map[*Node]chan *Event
}
