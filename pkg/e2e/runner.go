package e2e

import (
	"context"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	_ "net/http/pprof"
)

type Runner struct {
	ctx    context.Context
	log    *zap.Logger
	config *Config
	prom   *prometheus.Registry
	suite  *Suite
}

func NewRunner(ctx context.Context, log *zap.Logger, config *Config) *Runner {
	promReg := prometheus.NewRegistry()
	promReg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	promReg.MustRegister(collectors.NewGoCollector())
	return &Runner{
		ctx:    ctx,
		log:    log,
		config: config,
		prom:   promReg,
		suite:  NewSuite(ctx, log, config),
	}
}

func (r *Runner) Start() error {
	if r.config.Continuous {
		go func() {
			err := http.ListenAndServe("0.0.0.0:6061", nil)
			if err != nil {
				r.log.Error("serving profiler", zap.Error(err))
			}
		}()
	}

	return r.withMetricsServer(func() error {
		for {
			g, _ := errgroup.WithContext(r.ctx)
			for _, test := range r.suite.Tests() {
				test := test
				g.Go(func() error {
					return r.runTest(test)
				})
			}
			err := g.Wait()
			if err != nil && r.config.ContinuousExitOnError {
				return err
			}

			if !r.config.Continuous {
				break
			}
			time.Sleep(time.Duration(r.config.DelayBetweenRunsSeconds) * time.Second)
		}
		return nil
	})
}

func (r *Runner) runTest(test *Test) error {
	started := time.Now().UTC()
	log := r.log.With(zap.String("test", test.Name))

	err := test.Run(log)
	ended := time.Now().UTC()
	duration := ended.Sub(started)
	log = log.With(zap.Duration("duration", duration))
	if err != nil {
		recordFailedRun(r.ctx, test.Name)
		log.Error("test failed", zap.Error(err))
		recordRunDuration(r.ctx, duration, test.Name, "failed")
		return err
	}
	log.Info("test passed")

	recordSuccessfulRun(r.ctx, test.Name)
	recordRunDuration(r.ctx, duration, test.Name, "success")

	return nil
}
