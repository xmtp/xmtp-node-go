package ratelimiter

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/xmtp/xmtp-node-go/pkg/logging"
	"github.com/xmtp/xmtp-node-go/pkg/metrics"
	"go.uber.org/zap"
)

type LimitType string

const (
	DEFAULT_PRIORITY_MULTIPLIER      = uint16(5)
	PUBLISH_PRIORITY_MULTIPLIER      = uint16(25)
	DEFAULT_RATE_PER_MINUTE          = uint16(2000)
	DEFAULT_MAX_TOKENS               = uint16(10000)
	PUBLISH_RATE_PER_MINUTE          = uint16(200)
	PUBLISH_MAX_TOKENS               = uint16(1000)
	IDENTITY_PUBLISH_RATE_PER_MINUTE = uint16(10)
	IDENTITY_PUBLISH_MAX_TOKENS      = uint16(50)
	MAX_UINT_16                      = 65535

	DEFAULT          LimitType = "DEF"
	V3_DEFAULT       LimitType = "V3DEF"
	PUBLISH          LimitType = "PUB"
	V3_PUBLISH       LimitType = "V3PUB"
	IDENTITY_PUBLISH LimitType = "IDPUB"
)

type RateLimiter interface {
	Spend(limitType LimitType, bucket string, cost uint16, isPriority bool) error
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

// Limit controls token refilling for bucket entries
type Limit struct {
	// Maximum number of tokens that can be accumulated
	MaxTokens uint16
	// Number of tokens to refill per minute
	RatePerMinute uint16
}

func (l Limit) Refill(entry *Entry, multiplier uint16) {
	now := time.Now()
	ratePerMinute := l.RatePerMinute * multiplier
	maxTokens := l.MaxTokens * multiplier
	entry.mutex.Lock()
	defer entry.mutex.Unlock()
	minutesSinceLastSeen := now.Sub(entry.lastSeen).Minutes()
	if minutesSinceLastSeen > 0 {
		// Only update the lastSeen if it has been >= 1 minute
		// This allows for continuously sending nodes to still get credits
		entry.lastSeen = now
		// Convert to ints so that we can check if above MAX_UINT_16
		additionalTokens := int(ratePerMinute) * int(minutesSinceLastSeen)
		// Avoid overflows of UINT16 when new balance is above limit
		if additionalTokens+int(entry.tokens) > MAX_UINT_16 {
			additionalTokens = MAX_UINT_16 - int(entry.tokens)
		}
		entry.tokens = minUint16(entry.tokens+uint16(additionalTokens), maxTokens)
	}
}

// TokenBucketRateLimiter implements the RateLimiter interface
type TokenBucketRateLimiter struct {
	log                       *zap.Logger
	ctx                       context.Context
	mutex                     sync.RWMutex
	newBuckets                *Buckets // buckets that can be added to
	oldBuckets                *Buckets // buckets to be swept for expired entries
	PriorityMultiplier        uint16
	PublishPriorityMultiplier uint16
	Limits                    map[LimitType]*Limit
}

func NewTokenBucketRateLimiter(ctx context.Context, log *zap.Logger) *TokenBucketRateLimiter {
	tb := new(TokenBucketRateLimiter)
	tb.log = log.Named("ratelimiter")
	tb.ctx = ctx
	// TODO: need to periodically clear out expired items to avoid unlimited growth of the map.
	tb.newBuckets = NewBuckets(log, "buckets1")
	tb.oldBuckets = NewBuckets(log, "buckets2")
	tb.PriorityMultiplier = DEFAULT_PRIORITY_MULTIPLIER
	tb.PublishPriorityMultiplier = PUBLISH_PRIORITY_MULTIPLIER
	tb.Limits = map[LimitType]*Limit{
		DEFAULT:          {DEFAULT_MAX_TOKENS, DEFAULT_RATE_PER_MINUTE},
		V3_DEFAULT:       {DEFAULT_MAX_TOKENS, DEFAULT_RATE_PER_MINUTE},
		PUBLISH:          {PUBLISH_MAX_TOKENS, PUBLISH_RATE_PER_MINUTE},
		V3_PUBLISH:       {PUBLISH_MAX_TOKENS, PUBLISH_RATE_PER_MINUTE},
		IDENTITY_PUBLISH: {IDENTITY_PUBLISH_MAX_TOKENS, IDENTITY_PUBLISH_RATE_PER_MINUTE},
	}
	return tb
}

func (rl *TokenBucketRateLimiter) getLimit(limitType LimitType) *Limit {
	if l := rl.Limits[limitType]; l != nil {
		return l
	}
	return rl.Limits["default"]
}

// Will return the entry, with items filled based on the time since last access
func (rl *TokenBucketRateLimiter) fillAndReturnEntry(limitType LimitType, bucket string, isPriority bool) *Entry {
	limit := rl.getLimit(limitType)
	multiplier := uint16(1)
	if isPriority {
		// There are no priority limits for V3
		if limitType == PUBLISH {
			multiplier = rl.PublishPriorityMultiplier
		} else {
			multiplier = rl.PriorityMultiplier
		}
	}
	rl.mutex.RLock()
	if entry := rl.oldBuckets.getAndRefill(bucket, limit, multiplier, false); entry != nil {
		rl.mutex.RUnlock()
		return entry
	}
	entry := rl.newBuckets.getAndRefill(bucket, limit, multiplier, true)
	rl.mutex.RUnlock()
	return entry
}

// The Spend function takes a bucket and a boolean asserting whether to apply the PRIORITY or the REGULAR rate limits.
func (rl *TokenBucketRateLimiter) Spend(limitType LimitType, bucket string, cost uint16, isPriority bool) error {
	entry := rl.fillAndReturnEntry(limitType, bucket, isPriority)
	entry.mutex.Lock()
	defer entry.mutex.Unlock()
	log := rl.log.With(
		logging.String("bucket", bucket),
		logging.String("limitType", string(limitType)),
		logging.Bool("isPriority", isPriority),
		logging.Int("cost", int(cost)))
	if entry.tokens < cost {
		// Normally error strings should be fixed, but this error gets passed down to clients,
		// so we want to include more information for debugging purposes.
		// grpc Status has details in theory, but it seems messy to use, we may want to reconsider.
		return fmt.Errorf("%d exceeds rate limit %s", cost, bucket)
	}

	entry.tokens = entry.tokens - cost
	log.Debug("Spend allowed. bucket is under threshold", zap.Int("tokens_remaining", int(entry.tokens)))
	return nil
}

func (rl *TokenBucketRateLimiter) Janitor(sweepInterval, expiresAfter time.Duration) {
	ticker := time.NewTicker(sweepInterval)
	defer ticker.Stop()
	for {
		select {
		case <-rl.ctx.Done():
			return
		case <-ticker.C:
			rl.sweepAndSwap(expiresAfter)
		}
	}
}

func (rl *TokenBucketRateLimiter) sweepAndSwap(expiresAfter time.Duration) (deletedEntries int) {
	// Only the janitor writes to oldBuckets (the swap below), so we shouldn't need to rlock it here.
	deletedEntries = rl.oldBuckets.deleteExpired(expiresAfter)
	metrics.EmitRatelimiterDeletedEntries(rl.ctx, rl.oldBuckets.name, deletedEntries)
	metrics.EmitRatelimiterBucketsSize(rl.ctx, rl.oldBuckets.name, len(rl.oldBuckets.buckets))
	rl.mutex.Lock()
	rl.newBuckets, rl.oldBuckets = rl.oldBuckets, rl.newBuckets
	rl.mutex.Unlock()
	rl.newBuckets.log.Info("became new buckets")
	rl.oldBuckets.log.Info("became old buckets")
	return deletedEntries
}

func minUint16(x, y uint16) uint16 {
	if x <= y {
		return x
	}
	return y
}
