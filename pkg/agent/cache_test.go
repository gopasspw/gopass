package agent

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCache(t *testing.T) {
	testFactor := time.Duration(1)

	if value, ok := os.LookupEnv("SLOW_TEST_FACTOR"); ok {
		factor, err := strconv.Atoi(value)
		if err != nil {
			panic("Invalid SLOW_TEST_FACTOR set as environment variable")
		}
		testFactor = time.Duration(factor)
	}

	c := &cache{
		ttl:    10 * time.Millisecond * testFactor,
		maxTTL: 50 * time.Millisecond * testFactor,
	}

	val, found := c.get("foo")
	assert.Equal(t, "", val)
	assert.Equal(t, false, found)

	c.set("foo", "bar")
	val, found = c.get("foo")
	assert.Equal(t, "bar", val)
	assert.Equal(t, true, found)

	time.Sleep(5 * time.Millisecond * testFactor)
	val, found = c.get("foo")
	assert.Equal(t, "bar", val)
	assert.Equal(t, true, found)

	time.Sleep(12 * time.Millisecond * testFactor)
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
