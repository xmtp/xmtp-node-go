package history

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	SUCCESS_WALLET_ADDRESS = "0xgood"
	ERROR_WALLET_ADDRESS   = "0xerror"
)

type MockFetcher struct {
	NumFetches int
}

func NewMockFetcher() *MockFetcher {
	return new(MockFetcher)
}

func (m *MockFetcher) Fetch(ctx context.Context, walletAddress string) (res TransactionHistoryResult, err error) {
	m.NumFetches += 1
	if walletAddress == SUCCESS_WALLET_ADDRESS {
		return AlchemyTokenDataResult{hasTransactions: true}, nil
	}
	if walletAddress == ERROR_WALLET_ADDRESS {
		return nil, errors.New("Fetch error")
	}
	return res, err
}

func TestCacheSuccess(t *testing.T) {
	ctx := context.Background()
	mockFetcher := NewMockFetcher()
	fetcher := NewCacheingTransactionHistoryFetcher(mockFetcher)
	result, err := fetcher.Fetch(ctx, SUCCESS_WALLET_ADDRESS)
	require.NoError(t, err)
	require.Equal(t, result.HasTransactions(), true)
	require.Equal(t, mockFetcher.NumFetches, 1)

	_, err = fetcher.Fetch(ctx, SUCCESS_WALLET_ADDRESS)
	require.NoError(t, err)
	// Ensure that the underlying fetcher has only been called once
	require.Equal(t, mockFetcher.NumFetches, 1)
}

func TestCacheError(t *testing.T) {
	ctx := context.Background()
	mockFetcher := NewMockFetcher()
	fetcher := NewCacheingTransactionHistoryFetcher(mockFetcher)
	result, err := fetcher.Fetch(ctx, ERROR_WALLET_ADDRESS)
	require.Error(t, err)
	require.Nil(t, result)
	require.Equal(t, mockFetcher.NumFetches, 1)

	_, err = fetcher.Fetch(ctx, ERROR_WALLET_ADDRESS)
	require.Error(t, err)
	// Ensure that the underlying fetcher has been called twice since the first attempt failed
	require.Equal(t, mockFetcher.NumFetches, 2)
}

func TestRetrySuccess(t *testing.T) {
	ctx := context.Background()
	mockFetcher := NewMockFetcher()
	fetcher := NewRetryTransactionHistoryFetcher(mockFetcher, 5, 1*time.Second)
	result, err := fetcher.Fetch(ctx, SUCCESS_WALLET_ADDRESS)
	require.NoError(t, err)
	require.Equal(t, result.HasTransactions(), true)
	// Ensure that the mockFetcher was only called once
	require.Equal(t, mockFetcher.NumFetches, 1)
}

func TestRetryError(t *testing.T) {
	ctx := context.Background()
	mockFetcher := NewMockFetcher()
	fetcher := NewRetryTransactionHistoryFetcher(mockFetcher, 3, 1*time.Millisecond)
	result, err := fetcher.Fetch(ctx, ERROR_WALLET_ADDRESS)
	require.Error(t, err)
	require.Nil(t, result)
	// Ensure that the mockFetcher was called 3X
	require.Equal(t, mockFetcher.NumFetches, 3)
}

func TestRetryContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	mockFetcher := NewMockFetcher()
	fetcher := NewRetryTransactionHistoryFetcher(mockFetcher, 3, 500*time.Millisecond)
	var wg sync.WaitGroup
	var err error
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err = fetcher.Fetch(ctx, ERROR_WALLET_ADDRESS)
	}()
	time.Sleep(50 * time.Millisecond)
	// Cancel the context between the first and second iteration
	cancel()
	wg.Wait()
	require.Error(t, err)
	require.Equal(t, err.Error(), "context canceled")
	require.Equal(t, mockFetcher.NumFetches, 1)

}
