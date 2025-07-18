package prune

import (
	"context"
	"errors"

	"github.com/lib/pq"
	"github.com/xmtp/xmtp-node-go/pkg/mls/store/queries"
	"go.uber.org/zap"
)

const DEFAULT_GROUP_MESSAGE_LIFETIME_DAYS = 30

type GroupMessagesPruner struct {
	log       *zap.Logger
	querier   *queries.Queries
	batchSize int32
}

func (g *GroupMessagesPruner) Name() string {
	return "group_messages"
}

func NewGroupMessagesPruner(
	log *zap.Logger,
	querier *queries.Queries,
	batchSize int32,
) *GroupMessagesPruner {
	if batchSize < 100 {
		log.Panic("batchSize must be at least 100")
	}

	return &GroupMessagesPruner{
		log:       log,
		querier:   querier,
		batchSize: batchSize,
	}
}

func (g *GroupMessagesPruner) Count(ctx context.Context) (int64, error) {
	count, err := g.querier.CountDeletableGroupMessages(ctx, DEFAULT_GROUP_MESSAGE_LIFETIME_DAYS)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "57014" {
			g.log.Warn("Timeout while counting old group messages", zap.Error(err))
			return 1, nil
		}
		return 0, err
	}
	return count, nil
}

func (g *GroupMessagesPruner) PruneCycle(ctx context.Context) (int, error) {
	rows, err := g.querier.DeleteOldGroupMessagesBatch(
		ctx,
		queries.DeleteOldGroupMessagesBatchParams{
			AgeDays:   DEFAULT_GROUP_MESSAGE_LIFETIME_DAYS,
			BatchSize: g.batchSize,
		},
	)
	if err != nil {
		return 0, err
	}

	return len(rows), nil
}

var _ Pruner = (*GroupMessagesPruner)(nil)
