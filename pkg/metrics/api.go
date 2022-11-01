package metrics

import (
	"context"

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
