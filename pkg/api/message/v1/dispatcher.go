package api

import (
	"sync"

	bc "github.com/dustin/go-broadcast"
)

const (
	bufLen = 100
)

// dispatcher broadcasts messages to all channels registered for given topic.
type dispatcher struct {
	bcs  map[string]bc.Broadcaster
	subs map[string]map[chan interface{}]struct{}
	l    sync.RWMutex
}

func newDispatcher() *dispatcher {
	return &dispatcher{
		bcs:  map[string]bc.Broadcaster{},
		subs: map[string]map[chan interface{}]struct{}{},
	}
}

func (d *dispatcher) Register(topics ...string) chan interface{} {
	d.l.Lock()
	defer d.l.Unlock()
	ch := make(chan interface{})
	for _, topic := range topics {
		if _, exists := d.bcs[topic]; !exists {
			d.bcs[topic] = bc.NewBroadcaster(bufLen)
		}
		bc := d.bcs[topic]
		bc.Register(ch)
		if _, exists := d.subs[topic]; !exists {
			d.subs[topic] = map[chan interface{}]struct{}{}
		}
		d.subs[topic][ch] = struct{}{}
	}
	return ch
}

func (d *dispatcher) Unregister(ch chan interface{}, topics ...string) {
	d.l.Lock()
	defer d.l.Unlock()
	for _, topic := range topics {
		bc, exists := d.bcs[topic]
		if exists {
			bc.Unregister(ch)
		}
		if subs, exists := d.subs[topic]; exists {
			delete(subs, ch)
			if len(subs) == 0 && bc != nil {
				bc.Close()
				delete(d.bcs, topic)
			}
		}
	}
}

func (d *dispatcher) Submit(topic string, obj interface{}) bool {
	d.l.RLock()
	defer d.l.RUnlock()
	bc, exists := d.bcs[topic]
	if !exists {
		return false
	}
	return bc.TrySubmit(obj)
}

func (d *dispatcher) Close() error {
	d.l.Lock()
	defer d.l.Unlock()
	for topic, bc := range d.bcs {
		bc.Close()
		delete(d.bcs, topic)
	}
	return nil
}
