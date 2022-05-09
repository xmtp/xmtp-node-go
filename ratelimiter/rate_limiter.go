package ratelimiter

import (
	"errors"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"
)

const (
	ALLOW_LISTED_RATE_PER_MINUTE = uint16(10)
	ALLOW_LISTED_MAX_TOKENS      = uint16(10000)
	REGULAR_RATE_PER_MINUTE      = uint16(1)
	REGULAR_MAX_TOKENS           = uint16(100)
	MAX_UINT_16                  = 65535
)

type RateLimiter interface {
	Spend(walletAddress string, isAllowListed bool) error
}

// Entry represents a single wallet entry in the rate limiter
type Entry struct {
	// Time that the entry was last spent against. Updated at most once per minute
	LastSeen time.Time
	// This is more memory efficient but limits us to MaxTokens < 65535
	Tokens uint16
}

// TokenBucketRateLimiter implements the RateLimiter interface
type TokenBucketRateLimiter struct {
	log     *zap.Logger
	wallets map[string]*Entry
	mutex   sync.RWMutex
}

func NewTokenBucketRateLimiter(log *zap.Logger) *TokenBucketRateLimiter {
	tb := new(TokenBucketRateLimiter)
	tb.log = log.Named("ratelimiter")
	tb.wallets = make(map[string]*Entry)
	tb.mutex = sync.RWMutex{}

	return tb
}

func getRates(isAllowListed bool) (ratePerMinute uint16, maxTokens uint16) {
	if isAllowListed {
		ratePerMinute = ALLOW_LISTED_RATE_PER_MINUTE
		maxTokens = ALLOW_LISTED_MAX_TOKENS
	} else {
		ratePerMinute = REGULAR_RATE_PER_MINUTE
		maxTokens = REGULAR_MAX_TOKENS
	}
	return
}

// Will return the entry, with items filled based on the time since last access
func (rl *TokenBucketRateLimiter) fillAndReturnEntry(walletAddress string, isAllowListed bool) *Entry {
	ratePerMinute, maxTokens := getRates(isAllowListed)
	// The locking strategy is adapted from the following blog post: https://misfra.me/optimizing-concurrent-map-access-in-go/
	rl.mutex.RLock()
	currentVal, exists := rl.wallets[walletAddress]
	if !exists {
		// Unlock the read lock so we can get a new write lock to create a new entry
		rl.mutex.RUnlock()
		rl.mutex.Lock()
		currentVal = &Entry{
			Tokens:   uint16(maxTokens),
			LastSeen: time.Now(),
		}
		rl.wallets[walletAddress] = currentVal
		rl.mutex.Unlock()
	} else {
		rl.mutex.RUnlock()
	}

	now := time.Now()
	minutesSinceLastSeen := now.Sub(currentVal.LastSeen).Minutes()
	if minutesSinceLastSeen > 0 {
		// Only update the lastSeen if it has been >= 1 minute
		// This allows for continuously sending nodes to still get credits
		currentVal.LastSeen = now
		// Convert to ints so that we can check if above MAX_UINT_16
		additionalTokens := int(ratePerMinute) * int(minutesSinceLastSeen)
		// Avoid overflows of UINT16 when new balance is above limit
		if additionalTokens+int(currentVal.Tokens) > MAX_UINT_16 {
			additionalTokens = MAX_UINT_16 - int(currentVal.Tokens)
		}
		currentVal.Tokens = minUint16(currentVal.Tokens+uint16(additionalTokens), maxTokens)
	}

	return currentVal
}

// The Spend function takes a WalletAddress and a boolean asserting whether to apply the AllowListed rate limits or the regular rate limits
func (rl *TokenBucketRateLimiter) Spend(walletAddress string, isAllowListed bool) error {
	entry := rl.fillAndReturnEntry(walletAddress, isAllowListed)
	if entry.Tokens <= 0 {
		rl.log.Info("Rate limit exceeded", zap.String("wallet_address", walletAddress))
		return errors.New("rate_limit_exceeded")
	}

	entry.Tokens = entry.Tokens - 1
	rl.log.Debug("Spend allowed. Wallet is under threshold", zap.String("wallet_address", walletAddress), zap.Int("tokens_remaining", int(entry.Tokens)))

	return nil
}

func minUint16(x, y uint16) uint16 {
	return uint16(math.Min(float64(x), float64(y)))
}
