package metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	apicontext "github.com/xmtp/xmtp-node-go/pkg/api/message/v1/context"
	"github.com/xmtp/xmtp-node-go/pkg/logging"
	proto "github.com/xmtp/xmtp-node-go/pkg/proto/message_api/v1"
	"github.com/xmtp/xmtp-node-go/pkg/topic"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	appClientVersionTagKeys = []string{
		"client",
		"client_version",
		"app",
		"app_version",
	}
	apiRequestTagKeys = append([]string{
		"service",
		"method",
		"error_code",
	}, appClientVersionTagKeys...)

	apiRequestTagKeysByName = buildTagKeysMap(apiRequestTagKeys)

	topicCategoryTag   = "topic_category"
	queryParametersTag = "parameters"
	queryErrorTag      = "error"
)

var apiRequests = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "xmtp_api_requests",
		Help: "Count of api requests",
	},
	apiRequestTagKeys,
)

func EmitAPIRequest(ctx context.Context, log *zap.Logger, fields []zapcore.Field) {
	labels := prometheus.Labels{}
	for _, field := range fields {
		if !apiRequestTagKeysByName[field.Key] {
			continue
		}
		labels[field.Key] = field.String
	}
	for _, key := range apiRequestTagKeys {
		if _, ok := labels[key]; !ok {
			labels[key] = ""
		}
	}
	apiRequests.With(labels).Inc()
}

var subscribeTopicsLength = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "xmtp_subscribe_topics_length",
		Help:    "Number of subscribed topics per request",
		Buckets: []float64{1, 2, 4, 8, 16, 50, 100, 10000, 100000},
	},
	appClientVersionTagKeys,
)

func EmitSubscribeTopicsLength(ctx context.Context, log *zap.Logger, topics int) {
	labels := contextLabels(ctx)
	subscribeTopicsLength.With(labels).Observe(float64(topics))
}

var publishedEnvelopeSize = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "xmtp_published_envelope",
		Help:    "Size of a published envelope in bytes",
		Buckets: []float64{100, 1000, 10000, 100000},
	},
	append([]string{topicCategoryTag}, appClientVersionTagKeys...),
)

var publishedEnvelopeCount = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "xmtp_published_envelopes",
		Help: "Count of published envelopes",
	},
	append([]string{topicCategoryTag}, appClientVersionTagKeys...),
)

func EmitPublishedEnvelope(ctx context.Context, log *zap.Logger, env *proto.Envelope) {
	labels := contextLabels(ctx)
	topicCategory := topic.Category(env.ContentTopic)
	labels[topicCategoryTag] = topicCategory
	publishedEnvelopeSize.With(labels).Observe(float64(len(env.Message)))
	publishedEnvelopeCount.With(labels).Inc()
}

func contextLabels(ctx context.Context) prometheus.Labels {
	ri := apicontext.NewRequesterInfo(ctx)
	fields := ri.ZapFields()
	labels := prometheus.Labels{}
	for _, field := range fields {
		if !apiRequestTagKeysByName[field.Key] {
			continue
		}
		labels[field.Key] = field.String
	}
	return labels
}

func buildTagKeysMap(keys []string) map[string]bool {
	m := map[string]bool{}
	for _, key := range keys {
		m[key] = true
	}
	return m
}

var queryDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "xmtp_api_query_duration",
		Help:    "Duration of API query (ms)",
		Buckets: []float64{1, 10, 100, 1000, 10000, 100000},
	},
	append([]string{topicCategoryTag, queryErrorTag, queryParametersTag}, appClientVersionTagKeys...),
)

var queryResultLength = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "xmtp_api_query_result",
		Help:    "Number of events returned by an API query",
		Buckets: []float64{1, 10, 100, 1000, 10000, 100000},
	},
	append([]string{topicCategoryTag, queryErrorTag, queryParametersTag}, appClientVersionTagKeys...),
)

func EmitQuery(ctx context.Context, req *proto.QueryRequest, results int, err error, duration time.Duration) {
	labels := prometheus.Labels{}
	if len(req.ContentTopics) > 0 {
		labels[topicCategoryTag] = topic.Category(req.ContentTopics[0])
	}
	if err != nil {
		labels[queryErrorTag] = "internal"
	}
	parameters := logging.QueryShape(req)
	labels[queryParametersTag] = parameters

	queryDuration.With(labels).Observe(float64(duration.Milliseconds()))
	queryResultLength.With(labels).Observe(float64(results))
}
