package metrics

import (
	"context"
	"strings"

	proto "github.com/xmtp/proto/v3/go/message_api/v1"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	apiRequestTagKeys = []tag.Key{
		newTagKey("service"),
		newTagKey("method"),
		newTagKey("client"),
		newTagKey("client_version"),
		newTagKey("app"),
		newTagKey("app_version"),
		newTagKey("error_code"),
	}

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

func EmitAPIRequest(ctx context.Context, log *zap.Logger, fields []zapcore.Field) {
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
		log.Error("recording metric", fields...)
	}
}

var topicCategoryTag, _ = tag.NewKey("topic-category")
var publishedEnvelopeMeasure = stats.Int64("published_envelope", "size of a published envelope", stats.UnitBytes)
var publishedEnvelopeView = &view.View{
	Name:        "xmtp_published_envelope",
	Measure:     publishedEnvelopeMeasure,
	Description: "Size of a published envelope",
	Aggregation: view.Distribution(100, 1000, 10000, 100000),
	TagKeys:     []tag.Key{topicCategoryTag},
}

func EmitPublishedEnvelope(ctx context.Context, log *zap.Logger, env *proto.Envelope) {
	topicCategory := categoryFromTopic(env.ContentTopic)
	size := int64(len(env.Message))
	mutators := []tag.Mutator{tag.Insert(topicCategoryTag, topicCategory)}
	err := recordWithTags(ctx, mutators, publishedEnvelopeMeasure.M(size))
	if err != nil {
		log.Error("recording metric",
			zap.Error(err),
			zap.String("metric", publishedEnvelopeView.Name),
			zap.Int64("size", size),
			zap.String("topic_category", topicCategory),
		)
	}
}

var topicCategoryByPrefix = map[string]string{
	"test":         "test",
	"contact":      "contact",
	"intro":        "v1-intro",
	"dm":           "v1-conversation",
	"invite":       "v2-invite",
	"m":            "v2-conversation",
	"privatestore": "private",
}

func categoryFromTopic(contentTopic string) string {
	if strings.HasPrefix(contentTopic, "test-") {
		return "test"
	}
	topic := strings.TrimPrefix(contentTopic, "/xmtp/0/")
	if len(topic) == len(contentTopic) {
		return "invalid"
	}
	prefix, _, hasPrefix := strings.Cut(topic, "-")
	if hasPrefix {
		if category, found := topicCategoryByPrefix[prefix]; found {
			return category
		}
	}
	return "invalid"
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
