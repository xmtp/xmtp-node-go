package history

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type TransactionHistoryFetcher interface {
	Fetch(ctx context.Context, walletAddress string) (TransactionHistoryResult, error)
}

type TransactionHistoryResult interface {
	HasTransactions() bool
}

func NewDefaultTransactionHistoryFetcher(apiUrl string, log *zap.Logger) TransactionHistoryFetcher {
	return NewCacheingTransactionHistoryFetcher(
		NewRetryTransactionHistoryFetcher(
			NewAlchemyTransactionHistoryFetcher(apiUrl, log), 3, 1*time.Second,
		),
	)
}
