package e2e

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/xmtp/xmtp-node-go/pkg/metrics"
	"go.uber.org/zap"
)

var (
	testNameTagKey   = "test"
	testStatusTagKey = "status"

	successfulRuns = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "xmtpd_e2e_successful_runs",
			Help: "Number of successful runs",
		},
		[]string{testNameTagKey},
	)
	failedRuns = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "xmtpd_e2e_failed_runs",
			Help: "Number of failed runs",
		},
		[]string{testNameTagKey},
	)
	runDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "xmtpd_e2e_run_duration_seconds",
			Help:    "Duration of the run in seconds",
			Buckets: append(floatRange(30), 40, 50, 60, 90, 120, 300),
		},
		[]string{testNameTagKey, testStatusTagKey},
	)
)

func (r *Runner) withMetricsServer(fn func() error) error {
	metrics, err := metrics.NewMetricsServer(context.Background(), "0.0.0.0", 8008, r.log, r.prom)
	if err != nil {
		return err
	}
	defer func() {
		err := metrics.Close()
		if err != nil {
			r.log.Error("stopping metrics server", zap.Error(err))
		}
	}()

	r.prom.MustRegister(successfulRuns)
	r.prom.MustRegister(failedRuns)
	r.prom.MustRegister(runDuration)

	return fn()
}

func recordSuccessfulRun(ctx context.Context, testName string) {
	successfulRuns.WithLabelValues(testName).Inc()
}

func recordFailedRun(ctx context.Context, testName string) {
	failedRuns.WithLabelValues(testName).Inc()
}

func recordRunDuration(ctx context.Context, duration time.Duration, testName, testStatus string) {
	runDuration.WithLabelValues(testName, testStatus).Observe(duration.Seconds())
}

func floatRange(n int) []float64 {
	vals := make([]float64, n)
	for i := 0; i < n; i++ {
		vals[i] = float64(i + 1)
	}
	return vals
}
