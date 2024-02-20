package metrics

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	mlsstore "github.com/xmtp/xmtp-node-go/pkg/mls/store"
	"go.uber.org/zap"
)

var mlsSentGroupMessageSize = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "mls_sent_group_message_size",
		Help:    "Size of a sent group message in bytes",
		Buckets: []float64{100, 1000, 10000, 100000},
	},
	appClientVersionTagKeys,
)

var mlsSentGroupMessageCount = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "mls_sent_group_messages",
		Help: "Count of sent group messages",
	},
	appClientVersionTagKeys,
)

func EmitMLSSentGroupMessage(ctx context.Context, log *zap.Logger, msg *mlsstore.GroupMessage) {
	labels := contextLabels(ctx)
	mlsSentGroupMessageSize.With(labels).Observe(float64(len(msg.Data)))
	mlsSentGroupMessageCount.With(labels).Inc()
}

var mlsSentWelcomeMessageSize = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "mls_sent_welcome_message_size",
		Help:    "Size of a sent welcome message in bytes",
		Buckets: []float64{100, 1000, 10000, 100000},
	},
	appClientVersionTagKeys,
)

var mlsSentWelcomeMessageCount = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "mls_sent_welcome_messages",
		Help: "Count of sent welcome messages",
	},
	appClientVersionTagKeys,
)

func EmitMLSSentWelcomeMessage(ctx context.Context, log *zap.Logger, msg *mlsstore.WelcomeMessage) {
	labels := contextLabels(ctx)
	mlsSentWelcomeMessageSize.With(labels).Observe(float64(len(msg.Data)))
	mlsSentWelcomeMessageCount.With(labels).Inc()
}
