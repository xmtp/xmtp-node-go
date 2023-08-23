package metrics

import (
	"context"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
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
	recordWithTags(ctx, []tag.Mutator{tag.Insert(topicCategoryTag, name)}, ratelimiterBucketsGaugeMeasure.M(int64(size)))
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
	recordWithTags(ctx, []tag.Mutator{tag.Insert(topicCategoryTag, name)}, ratelimiterBucketsGaugeMeasure.M(int64(count)))
}
