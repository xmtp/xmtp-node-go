package metrics

import (
	"context"

	"github.com/xmtp/xmtp-node-go/pkg/logging"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.uber.org/zap"
)

var bucketsNameKey = newTagKey("name")

var ratelimiterBucketsGaugeMeasure = stats.Int64("ratelimiter_buckets", "size of ratelimiter buckets map", stats.UnitDimensionless)
var ratelimiterBucketsGaugeView = &view.View{
	Name:        "xmtp_ratelimiter_buckets",
	Measure:     ratelimiterBucketsGaugeMeasure,
	Description: "Size of rate-limiter buckets maps",
	Aggregation: view.LastValue(),
	TagKeys:     []tag.Key{bucketsNameKey},
}

func EmitRatelimiterBucketsSize(ctx context.Context, name string, size int) {
	err := recordWithTags(ctx, []tag.Mutator{tag.Insert(topicCategoryTag, name)}, ratelimiterBucketsGaugeMeasure.M(int64(size)))
	if err != nil {
		logging.From(ctx).Warn("recording metric",
			zap.String("metric", ratelimiterBucketsGaugeMeasure.Name()),
			zap.Error(err))
	}
}

var ratelimiterBucketsDeletedCounterMeasure = stats.Int64("xmtp_ratelimiter_entries_deleted", "Count of deleted entries from ratelimiter buckets map", stats.UnitDimensionless)
var ratelimiterBucketsDeletedCounterView = &view.View{
	Name:        "xmtp_ratelimiter_entries_deleted",
	Measure:     ratelimiterBucketsDeletedCounterMeasure,
	Description: "Count of deleted entries from rate-limiter buckets maps",
	Aggregation: view.Count(),
	TagKeys:     []tag.Key{bucketsNameKey},
}

func EmitRatelimiterDeletedEntries(ctx context.Context, name string, count int) {
	err := recordWithTags(ctx, []tag.Mutator{tag.Insert(topicCategoryTag, name)}, ratelimiterBucketsGaugeMeasure.M(int64(count)))
	if err != nil {
		logging.From(ctx).Warn("recording metric",
			zap.String("metric", ratelimiterBucketsDeletedCounterMeasure.Name()),
			zap.Error(err))
	}
}
