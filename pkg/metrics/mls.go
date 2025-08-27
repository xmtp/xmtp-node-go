package metrics

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/xmtp/xmtp-node-go/pkg/mls/store/queries"
	"go.uber.org/zap"
)

var sizeBuckets = []float64{100, 500, 1_000, 5_000, 10_000, 50_000, 100_000, 500_000, 1_000_000}

var mlsSentGroupMessageSize = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "mls_sent_group_message_size",
		Help:    "Size of a sent group message in bytes",
		Buckets: sizeBuckets,
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
		Buckets: sizeBuckets,
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
		Buckets: sizeBuckets,
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
	entry *queries.CommitLogV2,
) {
	labels := contextLabels(ctx)
	mlsCommitLogEntrySize.With(labels).
		Observe(float64(len(entry.SerializedEntry) + len(entry.SerializedSignature)))
	mlsCommitLogEntryCount.With(labels).Inc()
}

var mlsIdentityUpdateSize = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "mls_sent_identity_update_size",
		Help:    "Size of a sent identity update in bytes",
		Buckets: sizeBuckets,
	},
	appClientVersionTagKeys,
)

var mlsIdentityUpdateCount = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "mls_sent_identity_updates",
		Help: "Count of sent identity updates",
	},
	appClientVersionTagKeys,
)

func EmitMLSSentIdentityUpdate(ctx context.Context, log *zap.Logger, numBytes int) {
	labels := contextLabels(ctx)
	mlsIdentityUpdateSize.With(labels).Observe(float64(numBytes))
	mlsIdentityUpdateCount.With(labels).Inc()
}

var mlsKeyPackageSize = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "mls_sent_key_package_size",
		Help:    "Size of a key package in bytes",
		Buckets: sizeBuckets,
	},
	appClientVersionTagKeys,
)

var mlsKeyPackageCount = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "mls_sent_key_packages",
		Help: "Count of key packages",
	},
	appClientVersionTagKeys,
)

func EmitMLSPublishedKeyPackage(ctx context.Context, log *zap.Logger, numBytes int) {
	labels := contextLabels(ctx)
	mlsKeyPackageSize.With(labels).Observe(float64(numBytes))
	mlsKeyPackageCount.With(labels).Inc()
}
