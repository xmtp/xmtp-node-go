package metrics

import (
	"context"
	"strings"

	proto "github.com/xmtp/proto/go/message_api/v1"

	"github.com/xmtp/xmtp-node-go/pkg/logging"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.uber.org/zap"
)

var serviceNameTag, _ = tag.NewKey("service")
var methodNameTag, _ = tag.NewKey("method")
var clientNameTag, _ = tag.NewKey("client")
var clientVersionTag, _ = tag.NewKey("client-version")
var appNameTag, _ = tag.NewKey("app")
var appVersionTag, _ = tag.NewKey("app-version")

var apiRequestsMeasure = stats.Int64("api_requests", "Count api requests", stats.UnitDimensionless)

var apiRequestsView = &view.View{
	Name:        "xmtp_api_requests",
	Measure:     apiRequestsMeasure,
	Description: "Count of api requests",
	Aggregation: view.Count(),
	TagKeys: []tag.Key{
		serviceNameTag,
		methodNameTag,
		clientNameTag,
		clientVersionTag,
		appNameTag,
		appVersionTag,
	},
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

func EmitPublishedEnvelope(ctx context.Context, env *proto.Envelope) {
	topicCategory := categoryFromTopic(env.ContentTopic)
	size := int64(len(env.Message))
	mutators := []tag.Mutator{tag.Insert(topicCategoryTag, topicCategory)}
	err := recordWithTags(ctx, mutators, publishedEnvelopeMeasure.M(size))
	if err != nil {
		logging.From(ctx).Error("recording metric",
			zap.Error(err),
			zap.String("metric", publishedEnvelopeView.Name),
			zap.Int64("size", size),
			zap.String("topic_category", topicCategory),
		)
	}
}

func EmitAPIRequest(ctx context.Context, serviceName, methodName, clientName, clientVersion, appName, appVersion string) {
	mutators := []tag.Mutator{
		tag.Insert(clientNameTag, clientName),
		tag.Insert(clientVersionTag, clientVersion),
		tag.Insert(appNameTag, appName),
		tag.Insert(appVersionTag, appVersion),
		tag.Insert(serviceNameTag, serviceName),
		tag.Insert(methodNameTag, methodName),
	}
	err := recordWithTags(ctx, mutators, apiRequestsMeasure.M(1))
	if err != nil {
		logging.From(ctx).Error("recording metric",
			zap.Error(err),
			zap.String("service", serviceName),
			zap.String("method", methodName),
			zap.String("client", clientName),
			zap.String("client_version", clientVersion),
			zap.String("app", appName),
			zap.String("app_version", appVersion),
		)
	}
}

func categoryFromPrefix(topicPrefix string) string {
	for prefix, category := range map[string]string{
		"contact":      "contact",
		"intro":        "v1-intro",
		"dm":           "v1-conversation",
		"invite":       "v2-invite",
		"m":            "v2-conversation",
		"privatestore": "private",
	} {
		if topicPrefix == prefix {
			return category
		}
	}
	return "invalid"
}

func categoryFromTopic(topic string) string {
	prefix, _, hasPrefix := strings.Cut(topic, "-")
	if hasPrefix {
		return categoryFromPrefix(prefix)
	}
	return "invalid"
}
