package prune

import (
	"context"
	"errors"

	"github.com/lib/pq"
	"github.com/xmtp/xmtp-node-go/pkg/mls/store/queries"
	"go.uber.org/zap"
)

const DEFAULT_LIFETIME_OF_INSTALLATIONS = 90

type InstallationsPruner struct {
	log       *zap.Logger
	querier   *queries.Queries
	batchSize int32
}

func NewInstallationsPruner(log *zap.Logger, querier *queries.Queries, batchSize int32) *InstallationsPruner {
	if batchSize < 100 {
		log.Panic("batchSize must be at least 100")
	}

	return &InstallationsPruner{
		log:       log,
		querier:   querier,
		batchSize: batchSize,
	}
}

func (ip *InstallationsPruner) Count(ctx context.Context) (int64, error) {
	count, err := ip.querier.GetOldInstallations(ctx, DEFAULT_LIFETIME_OF_INSTALLATIONS)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "57014" {
			// timeout error
			// there might be millions of rows in the DB and a full table scan might take too long
			ip.log.Warn("Timeout error while counting old welcome messages", zap.Error(err))
			return 1, nil
		}
		return 0, err
	}
	return count, nil
}

func (ip *InstallationsPruner) PruneCycle(ctx context.Context) (int, error) {
	rows, err := ip.querier.DeleteOldInstallationsBatch(
		ctx,
		queries.DeleteOldInstallationsBatchParams{
			AgeDays:   DEFAULT_LIFETIME_OF_INSTALLATIONS,
			BatchSize: ip.batchSize,
		},
	)
	if err != nil {
		return 0, err
	}

	return len(rows), nil
}

var _ Pruner = (*InstallationsPruner)(nil)
