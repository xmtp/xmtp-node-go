package subscriptions

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"sync"

	proto "github.com/xmtp/xmtp-node-go/pkg/proto/message_api/v1"
	"github.com/xmtp/xmtp-node-go/pkg/topic"
	"go.uber.org/zap"
)

const (
	// allTopicsBacklogLength defines the buffer size for subscriptions that listen to all topics.
	allTopicsBacklogLength = 4096

	// minBacklogBufferLength defines the minimal length used for backlog buffer.
	minBacklogBufferLength = 10240

	WILDCARD_TOPIC = "*"

	v2TopicPrefix = "/xmtp/0/"
)

var ErrNilEnvelope = errors.New("received nil envelope")

// subscriptionDispatcher manages subscriptions and message dispatching.
type SubscriptionDispatcher struct {
	log           *zap.Logger                   // Logger instance
	subscriptions map[*Subscription]interface{} // Active subscriptions
	mu            sync.Mutex                    // Mutex for concurrency control
}

// newSubscriptionDispatcher creates a new dispatcher for managing subscriptions.
func NewSubscriptionDispatcher(log *zap.Logger) *SubscriptionDispatcher {
	dispatcher := &SubscriptionDispatcher{
		log:           log,
		subscriptions: make(map[*Subscription]interface{}),
	}

	return dispatcher
}

// messageHandler processes incoming messages, dispatching them to the correct subscription.
func (d *SubscriptionDispatcher) HandleEnvelope(env *proto.Envelope) {
	if env == nil {
		d.log.Warn("received nil envelope")
		return
	}

	xmtpTopic := isValidSubscribeAllTopic(env.ContentTopic)

	// Log message details for tracking
	d.log.Info("dispatching message to subscriptions",
		zap.String("content_topic", env.ContentTopic),
		zap.Uint64("timestamp_ns", env.TimestampNs),
		zap.String("message_digest", fmt.Sprintf("%x", computeDigest(env))),
		zap.Int("message_size_bytes", len(env.Message)),
	)

	d.mu.Lock()
	defer d.mu.Unlock()
	for subscription := range d.subscriptions {
		if subscription.all && !xmtpTopic {
			continue
		}
		if subscription.all || subscription.topics[env.ContentTopic] {
			select {
			case subscription.MessagesCh <- env:
				d.log.Debug("message dispatched to subscription",
					zap.String("content_topic", env.ContentTopic),
					zap.Uint64("timestamp_ns", env.TimestampNs),
					zap.String("message_digest", fmt.Sprintf("%x", computeDigest(env))),
					zap.Bool("is_subscribe_all", subscription.all),
					zap.Int("num_topics", len(subscription.topics)),
				)
			default:
				d.log.Info(
					fmt.Sprintf(
						"Subscription message channel is full, is subscribeAll: %t, numTopics: %d",
						subscription.all,
						len(subscription.topics),
					),
				)
				// we got here since the message channel was full. This happens when the client cannot
				// consume the data fast enough. In that case, we don't want to block further since it migth
				// slow down other users. Instead, we're going to close the channel and let the
				// consumer re-establish the connection if needed.
				close(subscription.MessagesCh)
				delete(d.subscriptions, subscription)
			}
		}
	}
}

// computeDigest calculates a digest for the envelope (same as in store package)
func computeDigest(env *proto.Envelope) []byte {
	digest := sha256.Sum256(append([]byte(env.ContentTopic), env.Message...))
	return digest[:]
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
func (d *SubscriptionDispatcher) Subscribe(topics map[string]bool) *Subscription {
	sub := &Subscription{
		dispatcher: d,
	}

	// Determine if subscribing to all topics or specific ones
	for topic := range topics {
		if WILDCARD_TOPIC == topic {
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
		sub.MessagesCh = make(chan *proto.Envelope, backlogBufferSize)
	} else {
		sub.MessagesCh = make(chan *proto.Envelope, allTopicsBacklogLength)
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	d.subscriptions[sub] = true
	return sub
}

func isValidSubscribeAllTopic(contentTopic string) bool {
	return strings.HasPrefix(contentTopic, v2TopicPrefix) || topic.IsMLSV1(contentTopic)
}
