package metrics

import (
	"context"
	"time"

	proto "github.com/xmtp/proto/v3/go/message_api/v1"
	apicontext "github.com/xmtp/xmtp-node-go/pkg/api/message/v1/context"
	"github.com/xmtp/xmtp-node-go/pkg/logging"
	"github.com/xmtp/xmtp-node-go/pkg/topic"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	appClientVersionTagKeys = []tag.Key{
		newTagKey("client"),
		newTagKey("client_version"),
		newTagKey("app"),
		newTagKey("app_version"),
	}
	apiRequestTagKeys = append([]tag.Key{
		newTagKey("service"),
		newTagKey("method"),
		newTagKey("error_code"),
	}, appClientVersionTagKeys...)

	apiRequestTagKeysByName = buildTagKeysByName(apiRequestTagKeys)
)

var apiRequestsMeasure = stats.Int64("api_requests", "Count api requests", stats.UnitDimensionless)

var apiRequestsView = &view.View{
	Name:        "xmtp_api_requests",
	Measure:     apiRequestsMeasure,
	Description: "Count of api requests",
	Aggregation: view.Count(),
	TagKeys:     apiRequestTagKeys,
}

func EmitAPIRequest(ctx context.Context, fields []zapcore.Field) {
	mutators := make([]tag.Mutator, 0, len(fields))
	for _, field := range fields {
		key, ok := apiRequestTagKeysByName[field.Key]
		if !ok {
			continue
		}
		mutators = append(mutators, tag.Insert(key, field.String))
	}
	err := recordWithTags(ctx, mutators, apiRequestsMeasure.M(1))
	if err != nil {
		logging.From(ctx).Error("recording metric", fields...)
	}
}

var topicCategoryTag, _ = tag.NewKey("topic-category")
var publishedEnvelopeMeasure = stats.Int64("published_envelope", "size of a published envelope", stats.UnitBytes)
var publishedEnvelopeView = &view.View{
	Name:        "xmtp_published_envelope",
	Measure:     publishedEnvelopeMeasure,
	Description: "Size of a published envelope",
	Aggregation: view.Distribution(100, 1000, 10000, 100000),
	TagKeys:     append([]tag.Key{topicCategoryTag}, appClientVersionTagKeys...),
}

var publishedEnvelopeCounterMeasure = stats.Int64("published_envelopes", "published envelopes", stats.UnitDimensionless)
var publishedEnvelopeCounterView = &view.View{
	Name:        "xmtp_published_envelopes",
	Measure:     publishedEnvelopeCounterMeasure,
	Description: "Count of published envelopes",
	Aggregation: view.Count(),
	TagKeys:     append([]tag.Key{topicCategoryTag}, appClientVersionTagKeys...),
}

func EmitPublishedEnvelope(ctx context.Context, env *proto.Envelope) {
	mutators := contextMutators(ctx)
	topicCategory := topic.Category(env.ContentTopic)
	mutators = append(mutators, tag.Insert(topicCategoryTag, topicCategory))
	size := int64(len(env.Message))
	err := recordWithTags(ctx, mutators, publishedEnvelopeMeasure.M(size))
	if err != nil {
		logging.From(ctx).Error("recording metric",
			zap.Error(err),
			zap.String("metric", publishedEnvelopeView.Name),
			zap.Int64("size", size),
			zap.String("topic_category", topicCategory),
		)
	}
	err = recordWithTags(ctx, mutators, publishedEnvelopeCounterMeasure.M(1))
	if err != nil {
		logging.From(ctx).Error("recording metric",
			zap.Error(err),
			zap.String("metric", publishedEnvelopeCounterView.Name),
			zap.Int64("size", size),
			zap.String("topic_category", topicCategory),
		)
	}
}

func contextMutators(ctx context.Context) []tag.Mutator {
	ri := apicontext.NewRequesterInfo(ctx)
	fields := ri.ZapFields()
	mutators := make([]tag.Mutator, 0, len(fields)+1)
	for _, field := range fields {
		key, ok := apiRequestTagKeysByName[field.Key]
		if !ok {
			continue
		}
		mutators = append(mutators, tag.Insert(key, field.String))
	}
	return mutators
}

func buildTagKeysByName(keys []tag.Key) map[string]tag.Key {
	m := map[string]tag.Key{}
	for _, key := range keys {
		m[key.Name()] = key
	}
	return m
}

func newTagKey(str string) tag.Key {
	key, _ := tag.NewKey(str)
	return key
}

var queryParametersTag, _ = tag.NewKey("parameters")
var queryErrorTag, _ = tag.NewKey("error")
var queryDurationMeasure = stats.Int64("api_query_duration", "duration of API query", stats.UnitMilliseconds)
var queryDurationView = &view.View{
	Name:        "xmtp_api_query_duration",
	Measure:     queryDurationMeasure,
	Description: "duration of API query (ms)",
	Aggregation: view.Distribution(1, 10, 100, 1000, 10000, 100000),
	TagKeys:     append([]tag.Key{topicCategoryTag, queryErrorTag, queryParametersTag}, appClientVersionTagKeys...),
}

var queryResultMeasure = stats.Int64("api_query_result", "number of events returned by an API query", stats.UnitNone)
var queryResultView = &view.View{
	Name:        "xmtp_api_query_result",
	Measure:     queryResultMeasure,
	Description: "number of events returned by an API query",
	Aggregation: view.Distribution(1, 10, 100, 1000, 10000, 100000),
	TagKeys:     append([]tag.Key{topicCategoryTag, queryErrorTag, queryParametersTag}, appClientVersionTagKeys...),
}

func EmitQuery(ctx context.Context, req *proto.QueryRequest, results int, err error, duration time.Duration) {
	mutators := contextMutators(ctx)
	if len(req.ContentTopics) > 0 {
		topicCategory := topic.Category(req.ContentTopics[0])
		mutators = append(mutators, tag.Insert(topicCategoryTag, topicCategory))
	}
	if err != nil {
		mutators = append(mutators, tag.Insert(queryErrorTag, "internal"))
	}
	parameters := logging.QueryShape(req)
	mutators = append(mutators, tag.Insert(queryParametersTag, parameters))
	err = recordWithTags(ctx, mutators, queryDurationMeasure.M(duration.Milliseconds()))
	if err != nil {
		logging.From(ctx).Error("recording metric",
			zap.Error(err),
			zap.Duration("duration", duration),
			zap.String("parameters", parameters),
			zap.String("metric", queryDurationView.Name),
		)
	}
	err = recordWithTags(ctx, mutators, queryResultMeasure.M(int64(results)))
	if err != nil {
		logging.From(ctx).Error("recording metric",
			zap.Error(err),
			zap.Int("results", results),
			zap.String("parameters", parameters),
			zap.String("metric", queryResultView.Name),
		)
	}
}
