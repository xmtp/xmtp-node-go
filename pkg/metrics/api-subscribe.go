package metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

var subscribeTopicsLength = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "xmtp_subscribe_topics_length",
		Help:    "Number of subscribed topics per request",
		Buckets: []float64{1, 2, 4, 8, 16, 50, 100, 10000, 100000},
	},
	appClientVersionTagKeys,
)

var subscribeTopics = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "xmtp_subscribe_topics",
		Help: "Count of subscribed topics",
	},
	appClientVersionTagKeys,
)

func EmitSubscribeTopics(ctx context.Context, log *zap.Logger, topics int) {
	labels := contextLabels(ctx)
	subscribeTopicsLength.With(labels).Observe(float64(topics))
	subscribeTopics.With(labels).Add(float64(topics))
}

var unsubscribeTopics = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "xmtp_unsubscribe_topics",
		Help: "Count of unsubscribed topics",
	},
	appClientVersionTagKeys,
)

func EmitUnsubscribeTopics(ctx context.Context, log *zap.Logger, topics int) {
	labels := contextLabels(ctx)
	unsubscribeTopics.With(labels).Add(float64(topics))
}

var unsubscribeTopicNatsDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "xmtp_unsubscribe_topic_nats_duration_ms",
		Help:    "Duration of unsubscribing from a topic via nats (in ms)",
		Buckets: []float64{1, 2, 4, 8, 16, 50, 100, 10000, 100000, 1000000},
	},
	appClientVersionTagKeys,
)

func EmitUnSubscribeTopicNatsDuration(ctx context.Context, log *zap.Logger, duration time.Duration) {
	labels := contextLabels(ctx)
	unsubscribeTopicNatsDuration.With(labels).Observe(float64(duration / time.Millisecond))
}
