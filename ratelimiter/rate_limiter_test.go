package ratelimiter

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const walletAddress = "0x1234"

func TestSpend(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	rl := NewTokenBucketRateLimiter(logger)
	rl.wallets[walletAddress] = &Entry{
		LastSeen: time.Now(),
		Tokens:   uint16(1),
	}

	err1 := rl.Spend(walletAddress, false)
	require.NoError(t, err1)
	err2 := rl.Spend(walletAddress, false)
	require.Error(t, err2)
	if err2.Error() != "rate_limit_exceeded" {
		t.Error("Incorrect error")
	}
}

// Ensure that new entries are created for previously unseen wallets
func TestSpendInitialize(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	rl := NewTokenBucketRateLimiter(logger)
	entry := rl.fillAndReturnEntry(walletAddress, false)
	require.Equal(t, entry.Tokens, REGULAR_MAX_TOKENS)
}

// Set the clock back 1 minute and ensure that 1 item has been added to the bucket
func TestSpendWithTime(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	rl := NewTokenBucketRateLimiter(logger)
	rl.wallets[walletAddress] = &Entry{
		// Set the last seen to 1 minute ago
		LastSeen: time.Now().Add(-1 * time.Minute),
		Tokens:   uint16(0),
	}
	err1 := rl.Spend(walletAddress, false)
	require.NoError(t, err1)
	err2 := rl.Spend(walletAddress, false)
	require.Error(t, err2)
}

// Ensure that the token balance cannot go above the max bucket size
func TestSpendMaxBucket(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	rl := NewTokenBucketRateLimiter(logger)
	rl.wallets[walletAddress] = &Entry{
		// Set last seen to 500 minutes ago
		LastSeen: time.Now().Add(-500 * time.Minute),
		Tokens:   uint16(0),
	}
	entry := rl.fillAndReturnEntry(walletAddress, false)
	require.Equal(t, entry.Tokens, REGULAR_MAX_TOKENS)
}

// Ensure that the allow list is being correctly applied
func TestSpendAllowListed(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	rl := NewTokenBucketRateLimiter(logger)
	rl.wallets[walletAddress] = &Entry{
		// Set last seen to 500 minutes ago
		LastSeen: time.Now().Add(-500 * time.Minute),
		Tokens:   uint16(0),
	}
	entry := rl.fillAndReturnEntry(walletAddress, true)
	require.Equal(t, entry.Tokens, uint16(500*ALLOW_LISTED_RATE_PER_MINUTE))
}

func TestMaxUint16(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	rl := NewTokenBucketRateLimiter(logger)
	rl.wallets[walletAddress] = &Entry{
		// Set last seen to 1 million minutes ago
		LastSeen: time.Now().Add(-1000000 * time.Minute),
		Tokens:   uint16(0),
	}

	entry := rl.fillAndReturnEntry(walletAddress, true)
	require.Equal(t, entry.Tokens, uint16(ALLOW_LISTED_MAX_TOKENS))
}

func TestSpendConcurrent(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	rl := NewTokenBucketRateLimiter(logger)
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = rl.Spend(walletAddress, false)
		}()
	}

	wg.Wait()

	entry := rl.fillAndReturnEntry(walletAddress, false)
	require.Equal(t, entry.Tokens, uint16(0))
}
