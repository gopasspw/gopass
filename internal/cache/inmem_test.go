package cache

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTTL(t *testing.T) {
	t.Parallel()

	testFactor := time.Duration(1)

	if value, ok := os.LookupEnv("SLOW_TEST_FACTOR"); ok {
		factor, err := strconv.Atoi(value)
		if err != nil {
			panic("Invalid SLOW_TEST_FACTOR set as environment variable")
		}
		testFactor = time.Duration(factor)
	}

	c := &InMemTTL[string, string]{
		ttl:    10 * time.Millisecond * testFactor,
		maxTTL: 50 * time.Millisecond * testFactor,
	}

	val, found := c.Get("foo")
	assert.Equal(t, "", val)
	assert.False(t, found)

	c.Set("foo", "bar")
	val, found = c.Get("foo")
	assert.Equal(t, "bar", val)
	assert.True(t, found)

	time.Sleep(5 * time.Millisecond * testFactor)
	val, found = c.Get("foo")
	assert.Equal(t, "bar", val)
	assert.True(t, found)

	time.Sleep(12 * time.Millisecond * testFactor)
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

	for i := 0; i < 32; i++ {
		for j := 0; j < 32; j++ {
			t.Run("set"+strconv.Itoa(i), func(t *testing.T) {
				t.Parallel()
				c.Set(i, i)
				time.Sleep(time.Millisecond)
				iv, found := c.Get(i)
				assert.True(t, found)
				assert.Equal(t, i, iv)
			})
		}
	}
}
