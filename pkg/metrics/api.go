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
var clientVersionTag, _ = tag.NewKey("client-version")

var apiRequestsMeasure = stats.Int64("api_requests", "Count api requests by client version", stats.UnitDimensionless)

var apiRequestsView = &view.View{
	Name:        "xmtp_api_requests",
	Measure:     apiRequestsMeasure,
	Description: "Count of api requests",
	Aggregation: view.Count(),
	TagKeys: []tag.Key{
		serviceNameTag,
		methodNameTag,
		clientVersionTag,
	},
}

func EmitAPIRequest(ctx context.Context, serviceName, methodName, clientVersion string) {
	mutators := []tag.Mutator{
		tag.Insert(clientVersionTag, clientVersion),
		tag.Insert(serviceNameTag, serviceName),
		tag.Insert(methodNameTag, methodName),
	}
	err := recordWithTags(ctx, mutators, apiRequestsMeasure.M(1))
	if err != nil {
		logging.From(ctx).Warn("recording metric",
			zap.Error(err),
			zap.String("service", serviceName),
			zap.String("method", methodName),
			zap.String("client_version", clientVersion),
		)
	}
}
