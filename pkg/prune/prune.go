package prune

import (
	"context"
	"database/sql"
	"sync"
	"sync/atomic"
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

	if e.config.CountDeletable {
		deletableCount := int64(0)
		for _, pruner := range pruners {
			prunerCount, err := pruner.Count(e.ctx)
			if err != nil {
				return err
			}
			deletableCount += prunerCount
		}

		e.log.Info("Count of envelopes eligible for pruning", zap.Int64("count", deletableCount))

		if deletableCount == 0 {
			e.log.Info("No envelopes found for pruning")
			return nil
		}
	}

	if e.config.DryRun {
		e.log.Info("Dry run mode enabled. Nothing to do")
		return nil
	}

	var totalDeletionCount int64 = 0

	cyclesCompleted := 0

	for {
		if cyclesCompleted >= e.config.MaxCycles {
			e.log.Warn("Reached maximum pruning cycles", zap.Int("maxCycles", e.config.MaxCycles))
			break
		}

		var (
			wg               sync.WaitGroup
			deletedThisCycle int64
			errCh            = make(chan error, len(pruners))
		)

		wg.Add(len(pruners))
		for _, pruner := range pruners {
			go func(p Pruner) {
				defer wg.Done()

				deleteCnt, err := p.PruneCycle(e.ctx)
				if err != nil {
					errCh <- err
					return
				}

				atomic.AddInt64(&deletedThisCycle, int64(deleteCnt))
				atomic.AddInt64(&totalDeletionCount, int64(deleteCnt))
			}(pruner)
		}

		wg.Wait()
		close(errCh)

		if err := <-errCh; err != nil {
			return err
		}

		if deletedThisCycle == 0 {
			break
		}

		e.log.Info("Pruned expired envelopes batch", zap.Int64("count", deletedThisCycle))
		cyclesCompleted++
	}

	if totalDeletionCount == 0 {
		e.log.Info("No expired envelopes found")
	}

	e.log.Info(
		"Done",
		zap.Int64("pruned count", totalDeletionCount),
		zap.Duration("elapsed", time.Since(start)),
	)

	return nil
}
