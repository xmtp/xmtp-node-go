package metrics

import (
	"github.com/status-im/go-waku/waku/metrics"
	"go.opencensus.io/stats/view"
	"go.uber.org/zap"
)

// Server wraps go-waku metrics server, so that we don't need to reference the go-waku package anywhere
type Server struct {
	*metrics.Server
}

func NewMetricsServer(address string, port int, logger *zap.Logger) *Server {
	return &Server{metrics.NewMetricsServer(address, port, logger)}
}

func RegisterViews(logger *zap.Logger) {
	if err := view.Register(
		PeersByProtoView,
		BootstrapPeersView,
	); err != nil {
		logger.Fatal("registering metrics views", zap.Error(err))
	}
}
