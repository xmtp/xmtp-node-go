package metrics

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/xmtp/xmtp-node-go/pkg/tracing"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.uber.org/zap"
)

type Server struct {
	log  *zap.Logger
	http *http.Server
}

func NewMetricsServer(log *zap.Logger, address string, port int) *Server {
	return &Server{
		log: log.Named("metrics"),
		http: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: promhttp.Handler(),
		},
	}
}

func (s *Server) Start(ctx context.Context) {
	go tracing.PanicWrap(ctx, "metrics server", func(_ context.Context) {
		s.log.Info("server stopped", zap.Error(s.http.ListenAndServe()))
	})
}

func (s *Server) Stop(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}

func RegisterViews(logger *zap.Logger) {
	if err := view.Register(
		PeersByProtoView,
		BootstrapPeersView,
		StoredMessageView,
		apiRequestsView,
		publishedEnvelopeView,
	); err != nil {
		logger.Fatal("registering metrics views", zap.Error(err))
	}
}

func record(ctx context.Context, measurement stats.Measurement) {
	stats.Record(ctx, measurement)
}

func recordWithTags(ctx context.Context, mutators []tag.Mutator, measurement stats.Measurement) error {
	return stats.RecordWithTags(ctx, mutators, measurement)
}
