package prune

import (
	"context"
	"database/sql"
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
		&WelcomePruner{log: e.log, querier: querier},
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

	totalDeletionCount := 0

	cyclesCompleted := 0

	for {
		if cyclesCompleted >= e.config.MaxCycles {
			e.log.Warn("Reached maximum pruning cycles", zap.Int("maxCycles", e.config.MaxCycles))
			break
		}

		deletedThisCycle := 0
		for _, pruner := range pruners {
			deleteCnt, err := pruner.PruneCycle(e.ctx)
			if err != nil {
				return err
			}
			deletedThisCycle += deleteCnt
			totalDeletionCount += deleteCnt
		}

		if deletedThisCycle == 0 {
			break
		}

		e.log.Info("Pruned expired envelopes batch", zap.Int("count", deletedThisCycle))
		cyclesCompleted++
	}

	if totalDeletionCount == 0 {
		e.log.Info("No expired envelopes found")
	}

	e.log.Info(
		"Done",
		zap.Int("pruned count", totalDeletionCount),
		zap.Duration("elapsed", time.Since(start)),
	)

	return nil
}
