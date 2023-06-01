package metrics

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/status-im/go-waku/waku/metrics"
	v2metrics "github.com/status-im/go-waku/waku/v2/metrics"
	"github.com/xmtp/xmtp-node-go/pkg/logging"
	"github.com/xmtp/xmtp-node-go/pkg/tracing"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.uber.org/zap"
)

// Server wraps go-waku metrics server, so that we don't need to reference the go-waku package anywhere
type Server struct {
	waku *metrics.Server
	http *http.Server
}

// re-export go-waku functions
var (
	RecordStoreError = v2metrics.RecordStoreError
	RecordMessage    = v2metrics.RecordMessage
)

func NewMetricsServer(address string, port int, logger *zap.Logger) *Server {
	return &Server{waku: metrics.NewMetricsServer(address, port, logger)}
}

func (s *Server) Start(ctx context.Context) {
	log := logging.From(ctx).Named("metrics")
	go tracing.PanicWrap(ctx, "waku metrics server", func(_ context.Context) { s.waku.Start() })
	s.http = &http.Server{Addr: ":8009", Handler: promhttp.Handler()}
	go tracing.PanicWrap(ctx, "metrics server", func(_ context.Context) {
		log.Info("server stopped", zap.Error(s.http.ListenAndServe()))
	})
}

func (s *Server) Stop(ctx context.Context) error {
	wErr := s.waku.Stop(ctx)
	err := s.http.Shutdown(ctx)
	if wErr != nil {
		return wErr
	}
	return err
}

func RegisterViews(logger *zap.Logger) {
	if err := view.Register(
		PeersByProtoView,
		BootstrapPeersView,
		StoredMessageView,
		apiRequestsView,
		publishedEnvelopeView,
		publishedEnvelopeCounterView,
		queryDurationView,
		queryResultView,
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
