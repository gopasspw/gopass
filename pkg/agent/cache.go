package agent

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

type cache struct {
	sync.Mutex
	ttl     time.Duration
	maxTTL  time.Duration
	entries map[string]cacheEntry
}

func (c *cache) get(key string) (string, bool) {
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

func (c *cache) purgeExpired() {
	for k, ce := range c.entries {
		if ce.isExpired() {
			delete(c.entries, k)
		}
	}
}

func (c *cache) set(key, value string) {
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

func (c *cache) remove(key string) {
	c.Lock()
	defer c.Unlock()

	delete(c.entries, key)
}

func (c *cache) purge() {
	c.Lock()
	defer c.Unlock()

	c.entries = make(map[string]cacheEntry, 10)
}
