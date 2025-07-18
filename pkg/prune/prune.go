package prune

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/xmtp/xmtp-node-go/pkg/mls/store/queries"
	"github.com/xmtp/xmtp-node-go/pkg/server"
	"go.uber.org/zap"
)

type Executor struct {
	ctx      context.Context
	log      *zap.Logger
	writerDB *sql.DB
	config   *server.PruneConfig
}

func NewPruneExecutor(
	ctx context.Context,
	log *zap.Logger,
	writerDB *sql.DB,
	config *server.PruneConfig,
) *Executor {
	return &Executor{
		ctx:      ctx,
		log:      log,
		writerDB: writerDB,
		config:   config,
	}
}

func (e *Executor) Run() error {
	querier := queries.New(e.writerDB)
	start := time.Now()

	pruners := []Pruner{
		NewWelcomePruner(e.log, querier, e.config.BatchSize),
		NewGroupMessagesPruner(e.log, querier, e.config.BatchSize),
		NewInstallationsPruner(e.log, querier, e.config.BatchSize),
	}

	var (
		wg    sync.WaitGroup
		errCh = make(chan error, len(pruners))
	)

	wg.Add(len(pruners))

	for _, pruner := range pruners {
		go func(pruner Pruner) {
			defer wg.Done()
			logger := e.log.Named(pruner.Name())
			if e.config.CountDeletable {
				deletableCount := int64(0)
				for _, pruner := range pruners {
					prunerCount, err := pruner.Count(e.ctx)
					if err != nil {
						logger.Error("Error counting envelopes for pruning", zap.Error(err))
						errCh <- err
						return
					}
					deletableCount += prunerCount
				}

				logger.Info(
					"Count of envelopes eligible for pruning",
					zap.Int64("count", deletableCount),
				)

				if deletableCount == 0 {
					logger.Info("No envelopes found for pruning")
					return
				}
			}

			if e.config.DryRun {
				logger.Info("Dry run mode enabled. Nothing to do")
				return
			}

			var totalDeletionCount int64 = 0
			cyclesCompleted := 0

			for {
				if cyclesCompleted >= e.config.MaxCycles {
					logger.Warn(
						"Reached maximum pruning cycles",
						zap.Int("maxCycles", e.config.MaxCycles),
					)
					break
				}

				deletedThisCycle, err := pruner.PruneCycle(e.ctx)
				if err != nil {
					logger.Error("Error pruning envelopes", zap.Error(err))
					errCh <- err
					return
				}

				totalDeletionCount += int64(deletedThisCycle)

				if deletedThisCycle == 0 {
					break
				}

				logger.Info("Pruned expired envelopes batch", zap.Int("count", deletedThisCycle))
				cyclesCompleted++
			}

			if totalDeletionCount == 0 {
				logger.Info("No expired envelopes found")
			}

			logger.Info(
				"Done",
				zap.Int64("pruned count", totalDeletionCount),
				zap.Duration("elapsed", time.Since(start)),
			)
		}(pruner)
	}

	wg.Wait()

	e.log.Info("All pruners finished")

	if err := <-errCh; err != nil {
		return err
	}

	return nil
}
