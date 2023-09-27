package metrics

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
)

var bucketsNameKey = "name"

var ratelimiterBuckets = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "xmtp_ratelimiter_buckets",
		Help: "Size of rate-limiter buckets maps",
	},
	[]string{bucketsNameKey},
)

func EmitRatelimiterBucketsSize(ctx context.Context, name string, size int) {
	ratelimiterBuckets.WithLabelValues(name).Set(float64(size))
}

var ratelimiterBucketsDeleted = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "xmtp_ratelimiter_entries_deleted",
		Help: "Count of deleted entries from rate-limiter buckets maps",
	},
	[]string{bucketsNameKey},
)

func EmitRatelimiterDeletedEntries(ctx context.Context, name string, count int) {
	ratelimiterBucketsDeleted.WithLabelValues(name).Set(float64(count))
}
