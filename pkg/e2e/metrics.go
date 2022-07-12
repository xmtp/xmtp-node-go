package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xmtp/xmtp-node-go/pkg/metrics"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	metricstag "go.opencensus.io/tag"
	"go.uber.org/zap"
)

var (
	successfulRuns     = stats.Int64("successful_runs", "Number of successful runs", stats.UnitDimensionless)
	failedRuns         = stats.Int64("failed_runs", "Number of failed runs", stats.UnitDimensionless)
	runDurationSeconds = stats.Float64("run_duration_seconds", "Duration of the run in seconds", stats.UnitSeconds)

	testNameTagKey = metricstag.MustNewKey("test")

	views = []*view.View{
		{
			Name:        "xmtpd_e2e_successful_runs",
			Measure:     successfulRuns,
			Description: "Number of successful runs",
			Aggregation: view.Count(),
			TagKeys:     []metricstag.Key{testNameTagKey},
		},
		{
			Name:        "xmtpd_e2e_failed_runs",
			Measure:     failedRuns,
			Description: "Number of failed runs",
			Aggregation: view.Count(),
			TagKeys:     []metricstag.Key{testNameTagKey},
		},
		{
			Name:        "xmtpd_e2e_run_duration_seconds",
			Measure:     runDurationSeconds,
			Description: "Duration of the run in seconds",
			Aggregation: view.Distribution(append(floatRange(30), 40, 50, 60, 90, 120, 300)...),
			TagKeys:     []metricstag.Key{testNameTagKey},
		},
	}
)

func withMetricsServer(t *testing.T, ctx context.Context, fn func(t *testing.T)) {
	log, err := zap.NewDevelopment()
	require.NoError(t, err)

	metrics := metrics.NewMetricsServer("0.0.0.0", 8008, log)
	go metrics.Start()
	defer func() {
		err := metrics.Stop(ctx)
		assert.NoError(t, err)
	}()

	err = view.Register(views...)
	require.NoError(t, err)

	fn(t)
}

func recordSuccessfulRun(ctx context.Context, tags ...tag) error {
	return recordWithTags(ctx, tags, successfulRuns.M(1))
}

func recordFailedRun(ctx context.Context, tags ...tag) error {
	return recordWithTags(ctx, tags, failedRuns.M(1))
}

func recordRunDuration(ctx context.Context, duration time.Duration, tags ...tag) error {
	return recordWithTags(ctx, tags, runDurationSeconds.M(duration.Seconds()))
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