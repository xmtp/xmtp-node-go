package history

import "context"

type TransactionHistoryFetcher interface {
	Fetch(ctx context.Context, walletAddress string) (TransactionHistoryResult, error)
}

type TransactionHistoryResult interface {
	HasTransactions() bool
}
