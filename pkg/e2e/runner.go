package e2e

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type Runner struct {
	ctx    context.Context
	log    *zap.Logger
	config *Config
	suite  *Suite
}

func NewRunner(ctx context.Context, log *zap.Logger, config *Config) *Runner {
	return &Runner{
		ctx:    ctx,
		log:    log,
		config: config,
		suite:  NewSuite(ctx, log, config),
	}
}

func (r *Runner) Start() error {
	if r.config.Continuous {
		go func() {
			err := http.ListenAndServe("0.0.0.0:6060", nil)
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
	nameTag := newTag(testNameTagKey, test.Name)
	started := time.Now().UTC()
	log := r.log.With(zap.String("test", test.Name))

	err := test.Run(log)
	ended := time.Now().UTC()
	duration := ended.Sub(started)
	log = log.With(zap.Duration("duration", duration))
	if err != nil {
		recordErr := recordFailedRun(r.ctx, nameTag)
		if recordErr != nil {
			log.Error("recording failed run metric", zap.Error(err))
		}
		log.Error("test failed", zap.Error(err))
		return err
	}
	log.Info("test passed")

	err = recordSuccessfulRun(r.ctx, nameTag)
	if err != nil {
		log.Error("recording successful run metric", zap.Error(err))
		return err
	}

	err = recordRunDuration(r.ctx, duration, nameTag)
	if err != nil {
		log.Error("recording run duration", zap.Error(err))
		return err
	}

	return nil
}
