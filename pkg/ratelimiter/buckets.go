package ratelimiter

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

type Buckets struct {
	log     *zap.Logger
	mutex   sync.RWMutex
	buckets map[string]*Entry
}

func NewBuckets(log *zap.Logger) *Buckets {
	return &Buckets{
		log:     log,
		buckets: make(map[string]*Entry),
		mutex:   sync.RWMutex{},
	}
}

func (b *Buckets) getAndRefill(bucket string, limit *Limit, multiplier uint16, create bool) *Entry {
	// The locking strategy is adapted from the following blog post: https://misfra.me/optimizing-concurrent-map-access-in-go/
	b.mutex.RLock()
	currentVal, exists := b.buckets[bucket]
	b.mutex.RUnlock()
	if !exists {
		if !create {
			return nil
		}
		b.mutex.Lock()
		currentVal, exists = b.buckets[bucket]
		if !exists {
			currentVal = &Entry{
				tokens:   uint16(limit.MaxTokens * multiplier),
				lastSeen: time.Now(),
				mutex:    sync.Mutex{},
			}
			b.buckets[bucket] = currentVal
			b.mutex.Unlock()

			return currentVal
		}
	}

	limit.Refill(currentVal, multiplier)
	return currentVal
}

func (b *Buckets) deleteExpired(expiresAfter time.Duration) {
	// Use RLock to iterate over the map
	// to allow concurrent reads
	b.mutex.RLock()
	var expired []string
	for bucket, entry := range b.buckets {
		if time.Since(entry.lastSeen) > expiresAfter {
			expired = append(expired, bucket)
		}
	}
	b.mutex.RUnlock()
	if len(expired) == 0 {
		return
	}
	b.log.Info("found expired buckets", zap.Int("count", len(expired)))
	// Use Lock for individual deletes to avoid prolonged
	// lockout for readers.
	count := 0
	for _, bucket := range expired {
		b.mutex.Lock()
		// check lastSeen again in case it was updated in the meantime.
		if entry, exists := b.buckets[bucket]; exists && time.Since(entry.lastSeen) > expiresAfter {
			delete(b.buckets, bucket)
			count++
		}
		b.mutex.Unlock()
	}
	b.log.Info("deleted expired buckets", zap.Int("count", count))
}
