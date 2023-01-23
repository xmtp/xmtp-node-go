package metrics

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/xmtp/xmtp-node-go/pkg/logging"
	"github.com/xmtp/xmtp-node-go/pkg/tracing"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.uber.org/zap"
)

type Server struct {
	http *http.Server
}

func NewMetricsServer(address string, port int, logger *zap.Logger) *Server {
	return &Server{}
}

func (s *Server) Start(ctx context.Context) {
	log := logging.From(ctx).Named("metrics")
	s.http = &http.Server{Addr: ":8009", Handler: promhttp.Handler()}
	go tracing.PanicWrap(ctx, "metrics server", func(_ context.Context) {
		log.Info("server stopped", zap.Error(s.http.ListenAndServe()))
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
