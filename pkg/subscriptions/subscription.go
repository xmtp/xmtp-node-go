package subscriptions

import proto "github.com/xmtp/xmtp-node-go/pkg/proto/message_api/v1"

// subscription represents a single subscription, including its message channel and topics.
type Subscription struct {
	MessagesCh chan *proto.Envelope    // Channel for receiving messages
	topics     map[string]bool         // Map of topics to subscribe to
	all        bool                    // Flag indicating subscription to all topics
	dispatcher *SubscriptionDispatcher // Parent dispatcher
}

// Unsubscribe removes the subscription from its dispatcher.
func (sub *Subscription) Unsubscribe() {
	sub.dispatcher.mu.Lock()
	defer sub.dispatcher.mu.Unlock()
	delete(sub.dispatcher.subscriptions, sub)
}

// Add grows the live topic set in place (XIP-83 mutate-in-place). It is O(len(topics))
// and safe to call concurrently with dispatch: HandleEnvelope reads sub.topics while
// holding the dispatcher mutex, so mutating under the same lock serializes the two.
// A no-op on a wildcard ("all") subscription.
func (sub *Subscription) Add(topics ...string) {
	sub.dispatcher.mu.Lock()
	defer sub.dispatcher.mu.Unlock()
	if sub.all {
		return
	}
	if sub.topics == nil {
		sub.topics = make(map[string]bool, len(topics))
	}
	for _, t := range topics {
		sub.topics[t] = true
	}
}

// Remove shrinks the live topic set in place. See Add for concurrency notes.
func (sub *Subscription) Remove(topics ...string) {
	sub.dispatcher.mu.Lock()
	defer sub.dispatcher.mu.Unlock()
	for _, t := range topics {
		delete(sub.topics, t)
	}
}
