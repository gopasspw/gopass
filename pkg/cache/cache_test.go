package cache

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTTL(t *testing.T) {
	testFactor := time.Duration(1)

	if value, ok := os.LookupEnv("SLOW_TEST_FACTOR"); ok {
		factor, err := strconv.Atoi(value)
		if err != nil {
			panic("Invalid SLOW_TEST_FACTOR set as environment variable")
		}
		testFactor = time.Duration(factor)
	}

	c := &TTL{
		ttl:    10 * time.Millisecond * testFactor,
		maxTTL: 50 * time.Millisecond * testFactor,
	}

	val, found := c.Get("foo")
	assert.Equal(t, "", val)
	assert.Equal(t, false, found)

	c.Set("foo", "bar")
	val, found = c.Get("foo")
	assert.Equal(t, "bar", val)
	assert.Equal(t, true, found)

	time.Sleep(5 * time.Millisecond * testFactor)
	val, found = c.Get("foo")
	assert.Equal(t, "bar", val)
	assert.Equal(t, true, found)

	time.Sleep(12 * time.Millisecond * testFactor)
	val, found = c.Get("foo")
	assert.Equal(t, "", val)
	assert.Equal(t, false, found)

	c.Set("bar", "baz")
	val, found = c.Get("bar")
	assert.Equal(t, "baz", val)
	assert.Equal(t, true, found)

	c.Remove("bar")
	val, found = c.Get("bar")
	assert.Equal(t, "", val)
	assert.Equal(t, false, found)

	c.Set("foo", "bar")
	c.Set("bar", "baz")
	val, found = c.Get("bar")
	assert.Equal(t, "baz", val)
	assert.Equal(t, true, found)

	c.Purge()
	val, found = c.Get("bar")
	assert.Equal(t, "", val)
	assert.Equal(t, false, found)
}
