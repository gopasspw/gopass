package cache

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func nowFunc(ns int) func() time.Time {
	return func() time.Time {
		return time.Date(2000, 1, 1, 1, 1, 1, ns, time.UTC)
	}
}

func TestTTL(t *testing.T) {
	t.Parallel()

	c := &InMemTTL[string, string]{
		ttl:    4,
		maxTTL: 5,
	}

	now = nowFunc(0)

	val, found := c.Get("foo")
	assert.Equal(t, "", val)
	assert.False(t, found)

	c.Set("foo", "bar")
	val, found = c.Get("foo")
	assert.Equal(t, "bar", val)
	assert.True(t, found)

	now = nowFunc(4)

	val, found = c.Get("foo")
	assert.Equal(t, "bar", val)
	assert.True(t, found)

	now = nowFunc(6)

	val, found = c.Get("foo")
	assert.Equal(t, "", val)
	assert.False(t, found)

	c.Set("bar", "baz")
	val, found = c.Get("bar")
	assert.Equal(t, "baz", val)
	assert.True(t, found)

	c.Remove("bar")
	val, found = c.Get("bar")
	assert.Equal(t, "", val)
	assert.False(t, found)

	c.Set("foo", "bar")
	c.Set("bar", "baz")
	val, found = c.Get("bar")
	assert.Equal(t, "baz", val)
	assert.True(t, found)

	c.Purge()
	val, found = c.Get("bar")
	assert.Equal(t, "", val)
	assert.False(t, found)
}

func TestPar(t *testing.T) {
	t.Parallel()

	c := NewInMemTTL[int, int](time.Minute, time.Minute)

	now = nowFunc(0)

	for i := 0; i < 32; i++ {
		for j := 0; j < 32; j++ {
			t.Run("set"+strconv.Itoa(i), func(t *testing.T) {
				t.Parallel()
				c.Set(i, i)
				iv, found := c.Get(i)
				assert.True(t, found)
				assert.Equal(t, i, iv)
			})
		}
	}
}
