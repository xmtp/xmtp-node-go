package prune

import (
	"context"
	"database/sql"
	"github.com/xmtp/xmtp-node-go/pkg/server"
	"go.uber.org/zap"
	"time"
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
	// querier := queries.New(e.writerDB)
	totalDeletionCount := 0
	start := time.Now()

	if e.config.DryRun {
		e.log.Info("Dry run mode enabled. Nothing to do")
		return nil
	}

	e.log.Info(
		"Done",
		zap.Int("pruned count", totalDeletionCount),
		zap.Duration("elapsed", time.Since(start)),
	)

	return nil
}
