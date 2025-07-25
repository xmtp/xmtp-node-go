package prune

import (
	"context"
	"errors"

	"github.com/lib/pq"
	"github.com/xmtp/xmtp-node-go/pkg/mls/store/queries"
	"go.uber.org/zap"
)

const DEFAULT_LIFETIME_OF_KEY_PACKAGES = 90

type KeyPackagesPruner struct {
	log       *zap.Logger
	querier   *queries.Queries
	batchSize int32
}

func (ip *KeyPackagesPruner) Name() string {
	return "key_packages"
}

func NewKeyPackagesPruner(
	log *zap.Logger,
	querier *queries.Queries,
	batchSize int32,
) *KeyPackagesPruner {
	if batchSize < 100 {
		log.Panic("batchSize must be at least 100")
	}

	return &KeyPackagesPruner{
		log:       log,
		querier:   querier,
		batchSize: batchSize,
	}
}

func (ip *KeyPackagesPruner) Count(ctx context.Context) (int64, error) {
	count, err := ip.querier.GetOldKeyPackages(ctx, DEFAULT_LIFETIME_OF_KEY_PACKAGES)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "57014" {
			// timeout error
			// there might be millions of rows in the DB and a full table scan might take too long
			ip.log.Warn("Timeout error while counting old key packages", zap.Error(err))
			return 1, nil
		}
		return 0, err
	}
	return count, nil
}

func (ip *KeyPackagesPruner) PruneCycle(ctx context.Context) (int, error) {
	rows, err := ip.querier.DeleteOldKeyPackagesBatch(
		ctx,
		queries.DeleteOldKeyPackagesBatchParams{
			AgeDays:   DEFAULT_LIFETIME_OF_KEY_PACKAGES,
			BatchSize: ip.batchSize,
		},
	)
	if err != nil {
		return 0, err
	}

	return len(rows), nil
}

var _ Pruner = (*KeyPackagesPruner)(nil)
