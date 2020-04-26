package cache

import (
	"sync"
	"time"
)

type cacheEntry struct {
	value     string
	maxExpire time.Time
	expire    time.Time
	created   time.Time
}

func (ce *cacheEntry) isExpired() bool {
	if time.Now().After(ce.maxExpire) {
		return true
	}
	if time.Now().After(ce.expire) {
		return true
	}
	return false
}

// TTL implements a simple TTLed cache. It is concurrency safe.
type TTL struct {
	sync.Mutex
	ttl     time.Duration
	maxTTL  time.Duration
	entries map[string]cacheEntry
}

// NewTTL creates a new TTLed cache.
func NewTTL(ttl time.Duration, maxTTL time.Duration) *TTL {
	return &TTL{
		ttl:    ttl,
		maxTTL: maxTTL,
	}
}

// Get retrieves a single entry, extending it's TTL.
func (c *TTL) Get(key string) (string, bool) {
	c.Lock()
	defer c.Unlock()

	if c.entries == nil {
		return "", false
	}

	ce, found := c.entries[key]
	if !found {
		// not found
		return "", false
	}
	if ce.isExpired() {
		// expired
		return "", false
	}
	ce.expire = time.Now().Add(c.ttl)
	c.entries[key] = ce
	return ce.value, true
}

// purgeExpire will remove expired entries. It is called by Set.
func (c *TTL) purgeExpired() {
	for k, ce := range c.entries {
		if ce.isExpired() {
			delete(c.entries, k)
		}
	}
}

// Set creates or overwrites an entry.
func (c *TTL) Set(key, value string) {
	c.Lock()
	defer c.Unlock()

	if c.entries == nil {
		c.entries = make(map[string]cacheEntry, 10)
	}

	now := time.Now()
	c.entries[key] = cacheEntry{
		value:     value,
		maxExpire: now.Add(c.maxTTL),
		expire:    now.Add(c.ttl),
		created:   now,
	}

	c.purgeExpired()
}

// Remove removes a single entry from the cache.
func (c *TTL) Remove(key string) {
	c.Lock()
	defer c.Unlock()

	delete(c.entries, key)
}

// Purge removes all entries from the cache.
func (c *TTL) Purge() {
	c.Lock()
	defer c.Unlock()

	c.entries = make(map[string]cacheEntry, 10)
}
