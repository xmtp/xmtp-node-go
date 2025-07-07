package prune

import (
	"context"
	"errors"
	"github.com/lib/pq"
	"github.com/xmtp/xmtp-node-go/pkg/mls/store/queries"
	"go.uber.org/zap"
)

const DEFAULT_LIFETIME_OF_WELCOME_MESSAGES = 90

type WelcomePruner struct {
	log     *zap.Logger
	querier *queries.Queries
}

func (w *WelcomePruner) Count(ctx context.Context) (int64, error) {
	count, err := w.querier.GetOldWelcomeMessages(ctx, DEFAULT_LIFETIME_OF_WELCOME_MESSAGES)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "57014" {
			// timeout error
			// there might be millions of rows in the DB and a full table scan might take too long
			w.log.Warn("Timeout error while counting old welcome messages", zap.Error(err))
			return 1, nil
		}
		return 0, err
	}
	return count, nil
}

func (w *WelcomePruner) PruneCycle(ctx context.Context) (int, error) {
	rows, err := w.querier.DeleteOldWelcomeMessagesBatch(ctx, DEFAULT_LIFETIME_OF_WELCOME_MESSAGES)
	if err != nil {
		return 0, err
	}

	return len(rows), nil
}

var _ Pruner = (*WelcomePruner)(nil)
