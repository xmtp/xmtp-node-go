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
