package api

import (
	"sync"

	gbc "github.com/dustin/go-broadcast"
)

const (
	bufLen = 100
)

// dispatcher broadcasts messages to all channels registered for given topic.
// A channel represents a subscription, it can be subscribed to multiple topics.
type dispatcher struct {
	// broadcaster for each topic
	bcsByTopic map[string]gbc.Broadcaster
	// sets of channels subscribed to each topic
	subsByTopic map[string]map[chan interface{}]bool
	// sets of subscribed topics for each channel
	topicsBySub map[chan interface{}]map[string]bool
	l           sync.RWMutex
}

func newDispatcher() *dispatcher {
	return &dispatcher{
		bcsByTopic:  make(map[string]gbc.Broadcaster),
		subsByTopic: make(map[string]map[chan interface{}]bool),
		topicsBySub: make(map[chan interface{}]map[string]bool),
	}
}

// Register updates subscriptions of ch to include topics.
// If ch is nil, it is created.
func (d *dispatcher) Register(ch chan interface{}, topics ...string) chan interface{} {
	if len(topics) == 0 {
		return nil
	}
	d.l.Lock()
	defer d.l.Unlock()
	if ch == nil {
		ch = make(chan interface{})
	}
	subTopics, exists := d.topicsBySub[ch]
	if !exists {
		subTopics = make(map[string]bool)
		d.topicsBySub[ch] = subTopics
	}
	for _, topic := range topics {
		if subTopics[topic] {
			continue // already subscribed
		}
		bc, exists := d.bcsByTopic[topic]
		if !exists {
			bc = gbc.NewBroadcaster(bufLen)
			d.bcsByTopic[topic] = bc
			d.subsByTopic[topic] = make(map[chan interface{}]bool)
		}
		bc.Register(ch)
		d.subsByTopic[topic][ch] = true
		subTopics[topic] = true
	}
	return ch
}

// Unregister updates the subscriptions of ch to exclude topics.
// If topics is empty, unsubscribe all current subscriptions of ch.
func (d *dispatcher) Unregister(ch chan interface{}, topics ...string) {
	d.l.Lock()
	defer d.l.Unlock()
	d.unregister(ch, topics...)
}

func (d *dispatcher) unregister(ch chan interface{}, topics ...string) {
	subTopics := d.topicsBySub[ch]
	if len(subTopics) == 0 {
		return
	}
	if len(topics) == 0 {
		// unsubscribe all current subscriptions
		for topic := range subTopics {
			topics = append(topics, topic)
		}
	}
	for _, topic := range topics {
		if !subTopics[topic] {
			continue
		}
		bc := d.bcsByTopic[topic]
		bc.Unregister(ch)
		subs := d.subsByTopic[topic]
		delete(subs, ch)
		if len(subs) == 0 {
			bc.Close()
			delete(d.bcsByTopic, topic)
			delete(d.subsByTopic, topic)
		}
		delete(subTopics, topic)
	}
	if len(subTopics) == 0 {
		delete(d.topicsBySub, ch)
	}
}

func (d *dispatcher) Submit(topic string, obj interface{}) bool {
	d.l.RLock()
	defer d.l.RUnlock()
	bc, exists := d.bcsByTopic[topic]
	if !exists {
		return false
	}
	return bc.TrySubmit(obj)
}

func (d *dispatcher) Close() error {
	d.l.Lock()
	defer d.l.Unlock()
	for ch := range d.topicsBySub {
		d.unregister(ch)
	}
	return nil
}
