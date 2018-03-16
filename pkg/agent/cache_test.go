package agent

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCache(t *testing.T) {
	c := &cache{
		ttl:    10 * time.Millisecond,
		maxTTL: 50 * time.Millisecond,
	}

	val, found := c.get("foo")
	assert.Equal(t, "", val)
	assert.Equal(t, false, found)

	c.set("foo", "bar")
	val, found = c.get("foo")
	assert.Equal(t, "bar", val)
	assert.Equal(t, true, found)

	time.Sleep(5 * time.Millisecond)
	val, found = c.get("foo")
	assert.Equal(t, "bar", val)
	assert.Equal(t, true, found)

	time.Sleep(12 * time.Millisecond)
	val, found = c.get("foo")
	assert.Equal(t, "", val)
	assert.Equal(t, false, found)

	c.set("bar", "baz")
	val, found = c.get("bar")
	assert.Equal(t, "baz", val)
	assert.Equal(t, true, found)

	c.remove("bar")
	val, found = c.get("bar")
	assert.Equal(t, "", val)
	assert.Equal(t, false, found)

	c.set("foo", "bar")
	c.set("bar", "baz")
	val, found = c.get("bar")
	assert.Equal(t, "baz", val)
	assert.Equal(t, true, found)

	c.purge()
	val, found = c.get("bar")
	assert.Equal(t, "", val)
	assert.Equal(t, false, found)
}
