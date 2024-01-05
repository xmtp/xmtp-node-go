package ratelimiter

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const walletAddress = "0x1234"

func TestSpend(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	rl := NewTokenBucketRateLimiter(context.Background(), logger)
	rl.newBuckets.getAndRefill(walletAddress, &Limit{1, 0}, 1, true)

	err1 := rl.Spend(DEFAULT, walletAddress, 1, false)
	require.NoError(t, err1)
	err2 := rl.Spend(DEFAULT, walletAddress, 1, false)
	require.Error(t, err2)
	require.Equal(t, "1 exceeds rate limit 0x1234", err2.Error())
}

// Ensure that new entries are created for previously unseen wallets
func TestSpendInitialize(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	rl := NewTokenBucketRateLimiter(context.Background(), logger)
	entry := rl.fillAndReturnEntry(DEFAULT, walletAddress, false)
	require.Equal(t, entry.tokens, DEFAULT_MAX_TOKENS)
}

// Set the clock back 1 minute and ensure that 1 item has been added to the bucket
func TestSpendWithTime(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	rl := NewTokenBucketRateLimiter(context.Background(), logger)
	rl.Limits[DEFAULT] = &Limit{100, 1}
	entry := rl.newBuckets.getAndRefill(walletAddress, &Limit{0, 0}, 1, true)
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
	rl := NewTokenBucketRateLimiter(context.Background(), logger)
	entry := rl.newBuckets.getAndRefill(walletAddress, &Limit{0, 0}, 1, true)
	// Set last seen to 500 minutes ago
	entry.lastSeen = time.Now().Add(-500 * time.Minute)
	entry = rl.fillAndReturnEntry(DEFAULT, walletAddress, false)
	require.Equal(t, entry.tokens, DEFAULT_MAX_TOKENS)
}

// Ensure that the allow list is being correctly applied
func TestSpendAllowListed(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	rl := NewTokenBucketRateLimiter(context.Background(), logger)
	entry := rl.newBuckets.getAndRefill(walletAddress, &Limit{0, 0}, 1, true)
	// Set last seen to 5 minutes ago
	entry.lastSeen = time.Now().Add(-5 * time.Minute)
	entry = rl.fillAndReturnEntry(DEFAULT, walletAddress, true)
	require.Equal(t, entry.tokens, uint16(5*DEFAULT_RATE_PER_MINUTE*DEFAULT_PRIORITY_MULTIPLIER))
}

func TestSpendAllowListedPublish(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	rl := NewTokenBucketRateLimiter(context.Background(), logger)
	entry := rl.newBuckets.getAndRefill(walletAddress, &Limit{0, 0}, 1, true)
	// Set last seen to 5 minutes ago
	entry.lastSeen = time.Now().Add(-5 * time.Minute)
	entry = rl.fillAndReturnEntry(PUBLISH, walletAddress, true)
	require.Equal(t, entry.tokens, uint16(5*PUBLISH_RATE_PER_MINUTE*PUBLISH_PRIORITY_MULTIPLIER))
}

func TestMaxUint16(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	rl := NewTokenBucketRateLimiter(context.Background(), logger)
	entry := rl.newBuckets.getAndRefill(walletAddress, &Limit{0, 0}, 1, true)
	// Set last seen to 1 million minutes ago
	entry.lastSeen = time.Now().Add(-1000000 * time.Minute)
	entry = rl.fillAndReturnEntry(DEFAULT, walletAddress, true)
	require.Equal(t, entry.tokens, DEFAULT_MAX_TOKENS*DEFAULT_PRIORITY_MULTIPLIER)
}

// Ensures that the map can be accessed concurrently
func TestSpendConcurrent(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	rl := NewTokenBucketRateLimiter(context.Background(), logger)
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

func TestBucketExpiration(t *testing.T) {
	// Set things up so that entries are expired after two sweep intervals
	expiresAfter := 100 * time.Millisecond
	sweepInterval := 60 * time.Millisecond

	logger, _ := zap.NewDevelopment()
	rl := NewTokenBucketRateLimiter(context.Background(), logger)
	rl.Limits[DEFAULT] = &Limit{2, 0} // 2 tokens, no refill

	require.NoError(t, rl.Spend(DEFAULT, "ip1", 1, false)) // bucket1 add
	require.NoError(t, rl.Spend(DEFAULT, "ip2", 1, false)) // bucket1 add

	time.Sleep(sweepInterval)
	require.Equal(t, 0, rl.sweepAndSwap(expiresAfter)) // sweep bucket2 and swap

	require.NoError(t, rl.Spend(DEFAULT, "ip2", 1, false)) // bucket1 refresh
	require.NoError(t, rl.Spend(DEFAULT, "ip3", 1, false)) // bucket2 add

	time.Sleep(sweepInterval)
	require.Equal(t, 1, rl.sweepAndSwap(expiresAfter)) // sweep bucket1 and swap, delete ip1

	// ip2 has been refreshed every 60ms so it should still be out of tokens
	require.Error(t, rl.Spend(DEFAULT, "ip2", 1, false)) // bucket1 refresh
	// ip1 entry should have expired by now, so we should have 2 tokens again
	require.NoError(t, rl.Spend(DEFAULT, "ip1", 1, false)) // bucket1 add
	require.NoError(t, rl.Spend(DEFAULT, "ip1", 1, false)) // bucket1 refresh
	require.Error(t, rl.Spend(DEFAULT, "ip1", 1, false))   // bucket1 refresh

	time.Sleep(sweepInterval)
	require.Equal(t, 1, rl.sweepAndSwap(expiresAfter)) // sweep bucket2 and swap, delete ip3

	// ip2 should still be out of tokens
	require.Error(t, rl.Spend(DEFAULT, "ip2", 1, false)) // bucket1 refresh
	// ip3 should have expired now and we should have 2 tokens again
	require.NoError(t, rl.Spend(DEFAULT, "ip3", 1, false)) // bucket2 add
	require.NoError(t, rl.Spend(DEFAULT, "ip3", 1, false)) // bucket2 refresh
	require.Error(t, rl.Spend(DEFAULT, "ip3", 1, false))   // bucket2 refresh
}

func TestBucketExpirationIntegrity(t *testing.T) {
	expiresAfter := 10 * time.Millisecond
	logger, _ := zap.NewDevelopment()
	rl := NewTokenBucketRateLimiter(context.Background(), logger)
	rl.Limits[DEFAULT] = &Limit{2, 0} // 2 tokens, no refill

	require.NoError(t, rl.Spend(DEFAULT, "ip1", 1, false)) // bucket1 add

	require.Equal(t, 0, rl.sweepAndSwap(expiresAfter)) // sweep bucket2 and swap

	require.NoError(t, rl.Spend(DEFAULT, "ip1", 1, false)) // bucket1 refresh
	require.Error(t, rl.Spend(DEFAULT, "ip1", 1, false))   // should be out of tokens now

	require.Equal(t, 0, rl.sweepAndSwap(expiresAfter)) // sweep bucket1 and swap

	require.Error(t, rl.Spend(DEFAULT, "ip1", 1, false)) // should still be out of tokens

	require.Equal(t, 0, rl.sweepAndSwap(expiresAfter)) // sweep bucket2 and swap

	time.Sleep(2 * expiresAfter)                       // wait until ip1 expires
	require.Equal(t, 1, rl.sweepAndSwap(expiresAfter)) // sweep bucket1 and swap, delete ip1

	require.NoError(t, rl.Spend(DEFAULT, "ip1", 1, false)) // bucket1 add
}
