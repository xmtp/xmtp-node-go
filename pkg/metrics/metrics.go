package metrics

import (
	"context"
	"fmt"
	"net/http"

	ocprometheus "contrib.go.opencensus.io/exporter/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
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

func NewMetricsServer(ctx context.Context, log *zap.Logger, address string, port int) (*Server, error) {
	exporter, err := ocprometheus.NewExporter(ocprometheus.Options{
		Registry: prometheus.DefaultRegisterer.(*prometheus.Registry),
	})
	if err != nil {
		return nil, err
	}

	// Badger DB metrics.
	// Based on https://grafana.com/grafana/dashboards/9574-badger/
	badgerExpvarCollector := collectors.NewExpvarCollector(map[string]*prometheus.Desc{
		"badger_blocked_puts_total":   prometheus.NewDesc("badger_blocked_puts_total", "Blocked Puts", nil, nil),
		"badger_disk_reads_total":     prometheus.NewDesc("badger_disk_reads_total", "Disk Reads", nil, nil),
		"badger_disk_writes_total":    prometheus.NewDesc("badger_disk_writes_total", "Disk Writes", nil, nil),
		"badger_gets_total":           prometheus.NewDesc("badger_gets_total", "Gets", nil, nil),
		"badger_puts_total":           prometheus.NewDesc("badger_puts_total", "Puts", nil, nil),
		"badger_memtable_gets_total":  prometheus.NewDesc("badger_memtable_gets_total", "Memtable gets", nil, nil),
		"badger_lsm_size_bytes":       prometheus.NewDesc("badger_lsm_size_bytes", "LSM Size in bytes", []string{"database"}, nil),
		"badger_vlog_size_bytes":      prometheus.NewDesc("badger_vlog_size_bytes", "Value Log Size in bytes", []string{"database"}, nil),
		"badger_pending_writes_total": prometheus.NewDesc("badger_pending_writes_total", "Pending Writes", []string{"database"}, nil),
		"badger_read_bytes":           prometheus.NewDesc("badger_read_bytes", "Read bytes", nil, nil),
		"badger_written_bytes":        prometheus.NewDesc("badger_written_bytes", "Written bytes", nil, nil),
		"badger_lsm_bloom_hits_total": prometheus.NewDesc("badger_lsm_bloom_hits_total", "LSM Bloom Hits", []string{"level"}, nil),
		"badger_lsm_level_gets_total": prometheus.NewDesc("badger_lsm_level_gets_total", "LSM Level Gets", []string{"level"}, nil),
	})
	err = prometheus.Register(badgerExpvarCollector)
	if err != nil {
		return nil, err
	}

	s := &Server{
		log: log.Named("metrics"),
		http: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: exporter,
		},
	}
	registerViews(log)
	s.Start(ctx)
	return s, nil
}

func (s *Server) Start(ctx context.Context) {
	go tracing.PanicWrap(ctx, "metrics server", func(_ context.Context) {
		s.log.Info("server stopped", zap.Error(s.http.ListenAndServe()))
	})
}

func (s *Server) Stop(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}

func registerViews(logger *zap.Logger) {
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
