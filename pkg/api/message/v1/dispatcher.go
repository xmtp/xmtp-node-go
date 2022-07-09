package api

import (
	"sync"

	bc "github.com/dustin/go-broadcast"
)

const (
	bufLen = 100
)

// dispatcher broadcasts messages to all channels
// registered for given content topic.
type dispatcher struct {
	m map[string]bc.Broadcaster
	l sync.RWMutex
}

func newDispatcher() *dispatcher {
	return &dispatcher{
		m: map[string]bc.Broadcaster{},
	}
}

func (b *dispatcher) Register(contentTopics ...string) chan interface{} {
	b.l.Lock()
	defer b.l.Unlock()
	ch := make(chan interface{})
	for _, contentTopic := range contentTopics {
		_, exists := b.m[contentTopic]
		if !exists {
			b.m[contentTopic] = bc.NewBroadcaster(bufLen)
		}
		bc := b.m[contentTopic]
		bc.Register(ch)
	}
	return ch
}

func (b *dispatcher) Unregister(ch chan interface{}, contentTopics ...string) {
	b.l.RLock()
	defer b.l.RUnlock()
	for _, contentTopic := range contentTopics {
		if bc, exists := b.m[contentTopic]; exists {
			bc.Unregister(ch)
		}
	}
}

func (b *dispatcher) Submit(contentTopic string, obj interface{}) bool {
	b.l.RLock()
	defer b.l.RUnlock()
	bc, exists := b.m[contentTopic]
	if !exists {
		return false
	}
	return bc.TrySubmit(obj)
}

func (b *dispatcher) Close() error {
	for _, bc := range b.m {
		bc.Close()
	}
	return nil
}
