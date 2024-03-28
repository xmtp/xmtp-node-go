package api

import (
	"fmt"
	"strings"
	"sync"

	"github.com/nats-io/nats.go"                                  // NATS messaging system
	proto "github.com/xmtp/xmtp-node-go/pkg/proto/message_api/v1" // Custom XMTP Protocol Buffers definition
	"github.com/xmtp/xmtp-node-go/pkg/topic"
	"go.uber.org/zap"                     // Logging library
	pb "google.golang.org/protobuf/proto" // Protocol Buffers for serialization
)

const (
	// allTopicsBacklogLength defines the buffer size for subscriptions that listen to all topics.
	allTopicsBacklogLength = 4096

	// minBacklogBufferLength defines the minimal length used for backlog buffer.
	minBacklogBufferLength = 1024
)

// subscriptionDispatcher manages subscriptions and message dispatching.
type subscriptionDispatcher struct {
	natsConn      *nats.Conn                    // Connection to NATS server
	natsSub       *nats.Subscription            // Subscription to NATS topics
	log           *zap.Logger                   // Logger instance
	subscriptions map[*subscription]interface{} // Active subscriptions
	mu            sync.Mutex                    // Mutex for concurrency control
}

// newSubscriptionDispatcher creates a new dispatcher for managing subscriptions.
func newSubscriptionDispatcher(conn *nats.Conn, log *zap.Logger) (*subscriptionDispatcher, error) {
	dispatcher := &subscriptionDispatcher{
		natsConn:      conn,
		log:           log,
		subscriptions: make(map[*subscription]interface{}),
	}

	// Subscribe to NATS wildcard topic and assign message handler
	var err error
	dispatcher.natsSub, err = conn.Subscribe(natsWildcardTopic, dispatcher.messageHandler)
	if err != nil {
		return nil, err
	}
	return dispatcher, nil
}

// Shutdown gracefully shuts down the dispatcher, unsubscribing from all topics.
func (d *subscriptionDispatcher) Shutdown() {
	_ = d.natsSub.Unsubscribe()
	// the lock/unlock ensures that there is no in-process dispatching.
	d.mu.Lock()
	defer d.mu.Unlock()
	d.natsSub = nil
	d.natsConn = nil
	d.subscriptions = nil
}

// messageHandler processes incoming messages, dispatching them to the correct subscription.
func (d *subscriptionDispatcher) messageHandler(msg *nats.Msg) {
	var env proto.Envelope
	err := pb.Unmarshal(msg.Data, &env)
	if err != nil {
		d.log.Info("unmarshaling envelope", zap.Error(err))
		return
	}
	d.log.Info("sending message to topic", zap.String("content_topic", env.ContentTopic))

	xmtpTopic := isValidSubscribeAllTopic(env.ContentTopic)

	d.mu.Lock()
	defer d.mu.Unlock()
	for subscription := range d.subscriptions {
		if subscription.all && !xmtpTopic {
			continue
		}
		if subscription.all || subscription.topics[env.ContentTopic] {
			select {
			case subscription.messagesCh <- &env:
			default:
				d.log.Info(fmt.Sprintf("Subscription message channel is full, is subscribeAll: %t, numTopics: %d", subscription.all, len(subscription.topics)))
				// we got here since the message channel was full. This happens when the client cannot
				// consume the data fast enough. In that case, we don't want to block further since it migth
				// slow down other users. Instead, we're going to close the channel and let the
				// consumer re-establish the connection if needed.
				// close(subscription.messagesCh)
				// delete(d.subscriptions, subscription)
			}
		}
	}
}

// subscription represents a single subscription, including its message channel and topics.
type subscription struct {
	messagesCh chan *proto.Envelope    // Channel for receiving messages
	topics     map[string]bool         // Map of topics to subscribe to
	all        bool                    // Flag indicating subscription to all topics
	dispatcher *subscriptionDispatcher // Parent dispatcher
}

// log2 calculates the base-2 logarithm of an integer using bitwise operations.
// It returns the floor of the actual base-2 logarithm.
func log2(n uint) (log2 uint) {
	if n == 0 {
		return 0
	}

	// Keep shifting n right until it becomes 0.
	// The number of shifts needed is the floor of log2(n).
	for n > 1 {
		n >>= 1
		log2++
	}
	return log2
}

// Subscribe creates a new subscription for the given topics.
func (d *subscriptionDispatcher) Subscribe(topics map[string]bool) *subscription {
	sub := &subscription{
		dispatcher: d,
	}

	// Determine if subscribing to all topics or specific ones
	for topic := range topics {
		if natsWildcardTopic == topic {
			sub.all = true
			break
		}
	}
	if !sub.all {
		sub.topics = topics
		// use a log2(length) as a backbuffer
		backlogBufferSize := log2(uint(len(topics))) + 1
		if backlogBufferSize < minBacklogBufferLength {
			backlogBufferSize = minBacklogBufferLength
		}
		sub.messagesCh = make(chan *proto.Envelope, backlogBufferSize)
	} else {
		sub.messagesCh = make(chan *proto.Envelope, allTopicsBacklogLength)
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	d.subscriptions[sub] = true
	return sub
}

// Unsubscribe removes the subscription from its dispatcher.
func (sub *subscription) Unsubscribe() {
	sub.dispatcher.mu.Lock()
	defer sub.dispatcher.mu.Unlock()
	delete(sub.dispatcher.subscriptions, sub)
}

func isValidSubscribeAllTopic(contentTopic string) bool {
	return strings.HasPrefix(contentTopic, validXMTPTopicPrefix) || topic.IsMLSV1(contentTopic)
}
