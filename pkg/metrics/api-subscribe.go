package metrics

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

var subscribedTopics = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "xmtp_subscribed_topics",
		Help: "Number of subscribed topics",
	},
	appClientVersionTagKeys,
)

var subscribeTopicsLength = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "xmtp_subscribe_topics_length",
		Help:    "Number of subscribed topics per request",
		Buckets: []float64{1, 2, 4, 8, 16, 50, 100, 10000, 100000},
	},
	appClientVersionTagKeys,
)

func EmitSubscribeTopics(ctx context.Context, log *zap.Logger, topics int) {
	labels := contextLabels(ctx)
	subscribeTopicsLength.With(labels).Observe(float64(topics))
	subscribedTopics.With(labels).Add(float64(topics))
}

func EmitUnsubscribeTopics(ctx context.Context, log *zap.Logger, topics int) {
	labels := contextLabels(ctx)
	subscribedTopics.With(labels).Add(-float64(topics))
}
