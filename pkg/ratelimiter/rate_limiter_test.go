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
	rl.buckets1.getAndRefill(walletAddress, &Limit{1, 0}, 1, true)

	err1 := rl.Spend(DEFAULT, walletAddress, 1, false)
	require.NoError(t, err1)
	err2 := rl.Spend(DEFAULT, walletAddress, 1, false)
	require.Error(t, err2)
	if err2.Error() != "rate limit exceeded" {
		t.Error("Incorrect error")
	}
}

// Ensure that new entries are created for previously unseen wallets
func TestSpendInitialize(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	rl := NewTokenBucketRateLimiter(logger)
	entry := rl.fillAndReturnEntry(DEFAULT, walletAddress, false)
	require.Equal(t, entry.tokens, DEFAULT_MAX_TOKENS)
}

// Set the clock back 1 minute and ensure that 1 item has been added to the bucket
func TestSpendWithTime(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	rl := NewTokenBucketRateLimiter(logger)
	rl.Limits[DEFAULT] = &Limit{100, 1}
	entry := rl.buckets1.getAndRefill(walletAddress, &Limit{0, 0}, 1, true)
	// Set the last seen to 1 minute ago
	entry.lastSeen = time.Now().Add(-1 * time.Minute)
	err1 := rl.Spend(DEFAULT, walletAddress, 1, false)
	require.NoError(t, err1)
	err2 := rl.Spend(DEFAULT, walletAddress, 1, false)
	require.Error(t, err2)
}

// Ensure that the token balance cannot go above the max bucket size
func TestSpendMaxBucket(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	rl := NewTokenBucketRateLimiter(logger)
	entry := rl.buckets1.getAndRefill(walletAddress, &Limit{0, 0}, 1, true)
	// Set last seen to 500 minutes ago
	entry.lastSeen = time.Now().Add(-500 * time.Minute)
	entry = rl.fillAndReturnEntry(DEFAULT, walletAddress, false)
	require.Equal(t, entry.tokens, DEFAULT_MAX_TOKENS)
}

// Ensure that the allow list is being correctly applied
func TestSpendAllowListed(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	rl := NewTokenBucketRateLimiter(logger)
	entry := rl.buckets1.getAndRefill(walletAddress, &Limit{0, 0}, 1, true)
	// Set last seen to 5 minutes ago
	entry.lastSeen = time.Now().Add(-5 * time.Minute)
	entry = rl.fillAndReturnEntry(DEFAULT, walletAddress, true)
	require.Equal(t, entry.tokens, uint16(5*DEFAULT_RATE_PER_MINUTE*PRIORITY_MULTIPLIER))
}

func TestMaxUint16(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	rl := NewTokenBucketRateLimiter(logger)
	entry := rl.buckets1.getAndRefill(walletAddress, &Limit{0, 0}, 1, true)
	// Set last seen to 1 million minutes ago
	entry.lastSeen = time.Now().Add(-1000000 * time.Minute)
	entry = rl.fillAndReturnEntry(DEFAULT, walletAddress, true)
	require.Equal(t, entry.tokens, DEFAULT_MAX_TOKENS*PRIORITY_MULTIPLIER)
}

// Ensures that the map can be accessed concurrently
func TestSpendConcurrent(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	rl := NewTokenBucketRateLimiter(logger)
	wg := sync.WaitGroup{}
	for i := 0; i < int(PUBLISH_MAX_TOKENS); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = rl.Spend(PUBLISH, walletAddress, 1, false)
		}()
	}

	wg.Wait()

	entry := rl.fillAndReturnEntry(PUBLISH, walletAddress, false)
	require.Equal(t, entry.tokens, uint16(0))
}
