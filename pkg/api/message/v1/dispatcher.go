package api

import (
	"sync"

	gbc "github.com/dustin/go-broadcast"
)

const (
	bufLen = 100
)

// dispatcher manages topic subscriptions.
// A channel represents a subscription, it can be subscribed to multiple topics.
// A broadcaster is used to broadcast messages from a given topic to all subscribed channels.
type dispatcher struct {
	// broadcaster for each topic
	bcsByTopic map[string]gbc.Broadcaster
	// sets of channels subscribed to each topic
	// parallels bcsByTopic, tracks which channels are subscribed to each topic
	subsByTopic map[string]map[chan interface{}]bool
	// sets of subscribed topics for each channel
	// this is the inverse of subsByTopic
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
		return nil // nothing to do
	}
	d.l.Lock()
	defer d.l.Unlock()
	if ch == nil {
		// create a channel if we weren't given one
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
			// new topic, set up a broadcaster
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
	if ch == nil {
		return
	}
	d.l.Lock()
	defer d.l.Unlock()
	d.unregister(ch, topics...)
}

// Update takes in an array of topics and unregisters any topics that are not in the array and registers any topics that are not already registered.
func (d *dispatcher) Update(ch chan interface{}, topics ...string) {
	if ch == nil {
		return
	}

	// Create a map of the new topics so we can check existing topics against it to see what needs to be added/removed
	newTopicMap := make(map[string]bool)
	for _, topic := range topics {
		newTopicMap[topic] = true
	}

	// Lock the map so we can check if any existing subscriptions need to be removed
	d.l.RLock()
	topicsBySub, hasTopicsBySub := d.topicsBySub[ch]
	toUnregister := make([]string, 0)
	for topic := range topicsBySub {
		if !newTopicMap[topic] {
			toUnregister = append(toUnregister, topic)
		}
	}
	d.l.RUnlock()

	// If the user is not subscribed to anything, just register everything in the list
	if !hasTopicsBySub {
		defer d.Register(ch, topics...)
		return
	}

	if len(toUnregister) > 0 {
		d.Unregister(ch, toUnregister...)
	}

	toRegister := make([]string, 0)

	// Lock again
	d.l.RLock()
	for _, topic := range topics {
		_, exists := topicsBySub[topic]
		if !exists {
			toRegister = append(toRegister, topic)
		}
	}
	d.l.RUnlock()
	d.Register(ch, toRegister...)

}

func (d *dispatcher) unregister(ch chan interface{}, topics ...string) {
	subTopics := d.topicsBySub[ch]
	if len(subTopics) == 0 {
		return // nothing to unsubscribe
	}
	if len(topics) == 0 {
		// unsubscribe all current subscriptions
		for topic := range subTopics {
			topics = append(topics, topic)
		}
	}
	for _, topic := range topics {
		if !subTopics[topic] {
			continue // not a subscribed topic
		}
		bc := d.bcsByTopic[topic]
		bc.Unregister(ch)
		subs := d.subsByTopic[topic]
		delete(subs, ch)
		if len(subs) == 0 {
			// no subscribers left, clean up topic
			bc.Close()
			delete(d.bcsByTopic, topic)
			delete(d.subsByTopic, topic)
		}
		delete(subTopics, topic)
	}
	if len(subTopics) == 0 {
		// no subscribed topics, drop the subscription
		delete(d.topicsBySub, ch)
	}
}

func (d *dispatcher) Submit(topic string, obj interface{}) bool {
	d.l.RLock()
	defer d.l.RUnlock()

	allBC, exists := d.bcsByTopic[contentTopicAllXMTP]
	if exists && isValidTopic(topic) {
		allBC.TrySubmit(obj)
	}

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
