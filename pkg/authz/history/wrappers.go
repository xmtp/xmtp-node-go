package history

import (
	"context"
	"time"

	lru "github.com/hashicorp/golang-lru"
)

const (
	CACHE_SIZE = 1000
)

// CacheingTransactionHistoryFetcher wraps a TransactionHistoryFetcher in a LRU cache
type CacheingTransactionHistoryFetcher struct {
	fetcher TransactionHistoryFetcher
	cache   *lru.TwoQueueCache
}

func NewCacheingTransactionHistoryFetcher(fetcher TransactionHistoryFetcher) TransactionHistoryFetcher {
	c := new(CacheingTransactionHistoryFetcher)
	c.fetcher = fetcher
	// We can safely swallow this error, since the only case it will blow up is if size == 0
	c.cache, _ = lru.New2Q(CACHE_SIZE)

	return c
}

func (f *CacheingTransactionHistoryFetcher) Fetch(ctx context.Context, walletAddress string) (res TransactionHistoryResult, err error) {
	if val, ok := f.cache.Get(walletAddress); ok {
		return val.(TransactionHistoryResult), nil
	}

	res, err = f.fetcher.Fetch(ctx, walletAddress)
	if err == nil {
		f.cache.Add(walletAddress, res)
	}
	return res, err
}

// RetryTransactionHistoryFetcher wraps a TransactionHistoryFetcher and will retry errors up to `numRetries` times with a sleep of `retrySleepTime` between each attempt
type RetryTransactionHistoryFetcher struct {
	fetcher        TransactionHistoryFetcher
	numRetries     int
	retrySleepTime time.Duration
}

func NewRetryTransactionHistoryFetcher(fetcher TransactionHistoryFetcher, numRetries int, retrySleepTime time.Duration) TransactionHistoryFetcher {
	res := new(RetryTransactionHistoryFetcher)
	res.fetcher = fetcher
	res.numRetries = numRetries
	res.retrySleepTime = retrySleepTime

	return res
}

func (f *RetryTransactionHistoryFetcher) Fetch(ctx context.Context, walletAddress string) (res TransactionHistoryResult, err error) {
	for i := 0; i < f.numRetries; i++ {
		select {
		case <-ctx.Done():
			return res, ctx.Err()
		default:
			res, err = f.fetcher.Fetch(ctx, walletAddress)
			if err == nil {
				return res, nil
			}
			time.Sleep(f.retrySleepTime)
		}
	}
	return res, err
}
