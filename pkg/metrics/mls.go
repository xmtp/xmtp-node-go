package metrics

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/xmtp/xmtp-node-go/pkg/mls/store/queries"
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

func EmitMLSSentGroupMessage(ctx context.Context, log *zap.Logger, msg *queries.GroupMessage) {
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

func EmitMLSSentWelcomeMessage(ctx context.Context, log *zap.Logger, msg *queries.WelcomeMessage) {
	labels := contextLabels(ctx)
	mlsSentWelcomeMessageSize.With(labels).Observe(float64(len(msg.Data)))
	mlsSentWelcomeMessageCount.With(labels).Inc()
}

var mlsCommitLogEntrySize = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "mls_commit_log_entry_size",
		Help:    "Size of a published commit log entry in bytes",
		Buckets: []float64{100, 1000, 10000, 100000},
	},
	appClientVersionTagKeys,
)

var mlsCommitLogEntryCount = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "mls_commit_log_entry_count",
		Help: "Count of published commit log entries",
	},
	appClientVersionTagKeys,
)

func EmitMLSPublishedCommitLogEntry(
	ctx context.Context,
	log *zap.Logger,
	entry *queries.CommitLog,
) {
	labels := contextLabels(ctx)
	mlsCommitLogEntrySize.With(labels).Observe(float64(len(entry.EncryptedEntry)))
	mlsCommitLogEntryCount.With(labels).Inc()
}
