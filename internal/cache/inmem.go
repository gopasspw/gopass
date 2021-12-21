package cache

import (
	"sync"
	"time"
)

type cacheEntry[V any] struct {
	value     V
	maxExpire time.Time
	expire    time.Time
	created   time.Time
}

func (ce *cacheEntry[V]) isExpired() bool {
	if time.Now().After(ce.maxExpire) {
		return true
	}
	if time.Now().After(ce.expire) {
		return true
	}
	return false
}

// InMemTTL implements a simple TTLed cache in memory. It is concurrency safe.
type InMemTTL[K comparable, V any] struct {
	sync.Mutex
	ttl     time.Duration
	maxTTL  time.Duration
	entries map[K]cacheEntry[V]
}

// NewInMemTTL creates a new TTLed cache.
func NewInMemTTL[K comparable, V any](ttl time.Duration, maxTTL time.Duration) *InMemTTL[K, V] {
	return &InMemTTL[K, V]{
		ttl:    ttl,
		maxTTL: maxTTL,
	}
}

// Get retrieves a single entry, extending it's TTL.
func (c *InMemTTL[K, V]) Get(key K) (V, bool) {
	c.Lock()
	defer c.Unlock()

	var zero V
	if c.entries == nil {
		return zero, false
	}

	ce, found := c.entries[key]
	if !found {
		// not found
		return zero, false
	}
	if ce.isExpired() {
		// expired
		return zero, false
	}

	ce.expire = time.Now().Add(c.ttl)
	c.entries[key] = ce
	return ce.value, true
}

// purgeExpire will remove expired entries. It is called by Set.
func (c *InMemTTL[K, V]) purgeExpired() {
	for k, ce := range c.entries {
		if ce.isExpired() {
			delete(c.entries, k)
		}
	}
}

// Set creates or overwrites an entry.
func (c *InMemTTL[K, V]) Set(key K, value V) {
	c.Lock()
	defer c.Unlock()

	if c.entries == nil {
		c.entries = make(map[K]cacheEntry[V], 10)
	}

	now := time.Now()
	c.entries[key] = cacheEntry[V]{
		value:     value,
		maxExpire: now.Add(c.maxTTL),
		expire:    now.Add(c.ttl),
		created:   now,
	}

	c.purgeExpired()
}

// Remove removes a single entry from the cache.
func (c *InMemTTL[K, V]) Remove(key K) {
	c.Lock()
	defer c.Unlock()

	delete(c.entries, key)
}

// Purge removes all entries from the cache.
func (c *InMemTTL[K, V]) Purge() {
	c.Lock()
	defer c.Unlock()

	c.entries = make(map[K]cacheEntry[V], 10)
}
