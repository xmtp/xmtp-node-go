package prune

import (
	"context"
	"github.com/xmtp/xmtp-node-go/pkg/mls/store/queries"
)

const DEFAULT_LIFETIME_OF_WELCOME_MESSAGES = 90

type WelcomePruner struct {
	querier *queries.Queries
}

func (w *WelcomePruner) Count(ctx context.Context) (int64, error) {
	return w.querier.GetOldWelcomeMessages(ctx, DEFAULT_LIFETIME_OF_WELCOME_MESSAGES)
}

func (w *WelcomePruner) PruneCycle(ctx context.Context) (int, error) {
	rows, err := w.querier.DeleteOldWelcomeMessagesBatch(ctx, DEFAULT_LIFETIME_OF_WELCOME_MESSAGES)
	if err != nil {
		return 0, err
	}

	return len(rows), nil
}

var _ Pruner = (*WelcomePruner)(nil)
