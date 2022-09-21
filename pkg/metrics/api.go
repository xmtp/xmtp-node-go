package metrics

import (
	"context"

	"github.com/xmtp/xmtp-node-go/pkg/logging"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.uber.org/zap"
)

var serviceNameTag, _ = tag.NewKey("service-name")
var methodNameTag, _ = tag.NewKey("method-name")
var clientVersionTag, _ = tag.NewKey("client-version")

var apiRequestsMeasure = stats.Int64("api_requests", "Count api requests by client version", stats.UnitDimensionless)

var apiRequestsView = &view.View{
	Name:        "xmtp_api_requests_by_client_version",
	Measure:     apiRequestsMeasure,
	Description: "Count of api requests by client version",
	Aggregation: view.Count(),
	TagKeys:     []tag.Key{clientVersionTag},
}

func EmitAPIRequest(ctx context.Context, serviceName, methodName, clientVersion string) {
	mutators := []tag.Mutator{
		tag.Insert(clientVersionTag, clientVersion),
		tag.Insert(serviceNameTag, serviceName),
		tag.Insert(methodNameTag, methodName),
	}
	err := stats.RecordWithTags(ctx, mutators)
	if err != nil {
		logging.From(ctx).Warn("recording metric",
			zap.String("service", serviceName),
			zap.String("method", methodName),
			zap.String("client_version", clientVersion),
		)
	}
}
