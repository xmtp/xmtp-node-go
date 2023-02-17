package e2e

import (
	"context"
	"time"

	"github.com/xmtp/xmtp-node-go/pkg/metrics"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	metricstag "go.opencensus.io/tag"
	"go.uber.org/zap"
)

var (
	runsMeasure               = stats.Int64("runs", "Number of runs", stats.UnitDimensionless)
	runDurationSecondsMeasure = stats.Float64("run_duration_seconds", "Duration of the run in seconds", stats.UnitSeconds)

	testNameTagKey   = metricstag.MustNewKey("test")
	testStatusTagKey = metricstag.MustNewKey("status")

	views = []*view.View{
		{
			Name:        "xmtpd_e2e_runs",
			Measure:     runsMeasure,
			Description: "Number of e2e runs",
			Aggregation: view.Count(),
			TagKeys:     []metricstag.Key{testNameTagKey, testStatusTagKey},
		},
		{
			Name:        "xmtpd_e2e_run_duration_seconds",
			Measure:     runDurationSecondsMeasure,
			Description: "Duration of the run in seconds",
			Aggregation: view.Distribution(append(floatRange(30), 40, 50, 60, 90, 120, 300)...),
			TagKeys:     []metricstag.Key{testNameTagKey},
		},
	}
)

func (r *Runner) withMetricsServer(fn func() error) error {
	metrics := metrics.NewMetricsServer("0.0.0.0", 8008, r.log)
	metrics.Start(context.Background())
	defer func() {
		err := metrics.Stop(r.ctx)
		if err != nil {
			r.log.Error("stopping metrics server", zap.Error(err))
		}
	}()

	err := view.Register(views...)
	if err != nil {
		return err
	}

	return fn()
}

func recordRun(ctx context.Context, tags ...tag) error {
	return recordWithTags(ctx, tags, runsMeasure.M(1))
}

func recordRunDuration(ctx context.Context, duration time.Duration, tags ...tag) error {
	return recordWithTags(ctx, tags, runDurationSecondsMeasure.M(duration.Seconds()))
}

type tag struct {
	key   metricstag.Key
	value string
}

func newTag(key metricstag.Key, value string) tag {
	return tag{
		key:   key,
		value: value,
	}
}

func recordWithTags(ctx context.Context, tags []tag, ms ...stats.Measurement) error {
	mutators := make([]metricstag.Mutator, len(tags))
	for i, tag := range tags {
		mutators[i] = metricstag.Upsert(tag.key, tag.value)
	}
	return stats.RecordWithTags(ctx, mutators, ms...)
}

func floatRange(n int) []float64 {
	vals := make([]float64, n)
	for i := 0; i < n; i++ {
		vals[i] = float64(i + 1)
	}
	return vals
}
