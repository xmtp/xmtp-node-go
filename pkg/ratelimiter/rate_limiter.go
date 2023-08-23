package ratelimiter

import (
	"errors"
	"sync"
	"time"

	"github.com/xmtp/xmtp-node-go/pkg/logging"
	"go.uber.org/zap"
)

const (
	PRIORITY_RATE_PER_MINUTE = uint16(10)
	PRIORITY_MAX_TOKENS      = uint16(10000)
	REGULAR_RATE_PER_MINUTE  = uint16(1)
	REGULAR_MAX_TOKENS       = uint16(100)
	MAX_UINT_16              = 65535
)

type RateLimiter interface {
	Spend(bucket string, cost uint16, isAllowListed bool) error
}

// Entry represents a single wallet entry in the rate limiter
type Entry struct {
	// Time that the entry was last spent against. Updated at most once per minute
	lastSeen time.Time
	// This is more memory efficient but limits us to MaxTokens < 65535
	tokens uint16
	// Add a per-entry mutex to avoid lock contention on the list as a whole
	mutex sync.Mutex
}

// TokenBucketRateLimiter implements the RateLimiter interface
type TokenBucketRateLimiter struct {
	log     *zap.Logger
	buckets map[string]*Entry
	mutex   sync.RWMutex
}

func NewTokenBucketRateLimiter(log *zap.Logger) *TokenBucketRateLimiter {
	tb := new(TokenBucketRateLimiter)
	tb.log = log.Named("ratelimiter")
	tb.buckets = make(map[string]*Entry)
	tb.mutex = sync.RWMutex{}
	return tb
}

func getRates(isPriority bool) (ratePerMinute uint16, maxTokens uint16) {
	if isPriority {
		ratePerMinute = PRIORITY_RATE_PER_MINUTE
		maxTokens = PRIORITY_MAX_TOKENS
	} else {
		ratePerMinute = REGULAR_RATE_PER_MINUTE
		maxTokens = REGULAR_MAX_TOKENS
	}
	return
}

// Will return the entry, with items filled based on the time since last access
func (rl *TokenBucketRateLimiter) fillAndReturnEntry(bucket string, isPriority bool) *Entry {
	ratePerMinute, maxTokens := getRates(isPriority)
	// The locking strategy is adapted from the following blog post: https://misfra.me/optimizing-concurrent-map-access-in-go/
	rl.mutex.RLock()
	currentVal, exists := rl.buckets[bucket]
	rl.mutex.RUnlock()
	if !exists {
		rl.mutex.Lock()
		currentVal = &Entry{
			tokens:   uint16(maxTokens),
			lastSeen: time.Now(),
			mutex:    sync.Mutex{},
		}
		rl.buckets[bucket] = currentVal
		rl.mutex.Unlock()

		return currentVal
	}

	currentVal.mutex.Lock()
	defer currentVal.mutex.Unlock()
	now := time.Now()
	minutesSinceLastSeen := now.Sub(currentVal.lastSeen).Minutes()
	if minutesSinceLastSeen > 0 {
		// Only update the lastSeen if it has been >= 1 minute
		// This allows for continuously sending nodes to still get credits
		currentVal.lastSeen = now
		// Convert to ints so that we can check if above MAX_UINT_16
		additionalTokens := int(ratePerMinute) * int(minutesSinceLastSeen)
		// Avoid overflows of UINT16 when new balance is above limit
		if additionalTokens+int(currentVal.tokens) > MAX_UINT_16 {
			additionalTokens = MAX_UINT_16 - int(currentVal.tokens)
		}
		currentVal.tokens = minUint16(currentVal.tokens+uint16(additionalTokens), maxTokens)
	}

	return currentVal
}

// The Spend function takes a bucket and a boolean asserting whether to apply the PRIORITY or the REGULAR rate limits.
func (rl *TokenBucketRateLimiter) Spend(bucket string, cost uint16, isPriority bool) error {
	entry := rl.fillAndReturnEntry(bucket, isPriority)
	entry.mutex.Lock()
	defer entry.mutex.Unlock()
	log := rl.log.With(logging.String("bucket", bucket))
	if entry.tokens < cost {
		log.Info("Rate limit exceeded")
		return errors.New("rate_limit_exceeded")
	}

	entry.tokens = entry.tokens - cost
	log.Debug("Spend allowed. bucket is under threshold", zap.Int("tokens_remaining", int(entry.tokens)))
	return nil
}

func minUint16(x, y uint16) uint16 {
	if x <= y {
		return x
	}
	return y
}
