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

// InMemTTL implements a simple TTLed cache in memory. It is concurrency safe.
type InMemTTL struct {
	sync.Mutex
	ttl     time.Duration
	maxTTL  time.Duration
	entries map[string]cacheEntry
}

// NewInMemTTL creates a new TTLed cache.
func NewInMemTTL(ttl time.Duration, maxTTL time.Duration) *InMemTTL {
	return &InMemTTL{
		ttl:    ttl,
		maxTTL: maxTTL,
	}
}

// Get retrieves a single entry, extending it's TTL.
func (c *InMemTTL) Get(key string) (string, bool) {
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
func (c *InMemTTL) purgeExpired() {
	for k, ce := range c.entries {
		if ce.isExpired() {
			delete(c.entries, k)
		}
	}
}

// Set creates or overwrites an entry.
func (c *InMemTTL) Set(key, value string) {
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
func (c *InMemTTL) Remove(key string) {
	c.Lock()
	defer c.Unlock()

	delete(c.entries, key)
}

// Purge removes all entries from the cache.
func (c *InMemTTL) Purge() {
	c.Lock()
	defer c.Unlock()

	c.entries = make(map[string]cacheEntry, 10)
}
