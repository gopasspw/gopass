package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOnDisk(t *testing.T) {
	t.Parallel()

	td := t.TempDir()

	odc, err := NewOnDiskWithDir("test", td, time.Hour)
	assert.NoError(t, err)

	assert.NoError(t, odc.Set("foo", []string{"bar"}))
	res, err := odc.Get("foo")
	assert.NoError(t, err)
	assert.Equal(t, []string{"bar"}, res)

	assert.Error(t, odc.Remove("bar"))
	assert.NoError(t, odc.Remove("foo"))
	assert.NoError(t, odc.Purge())
}

func TestOnDiskExpiry(t *testing.T) {
	t.Parallel()

	td := t.TempDir()

	odc, err := NewOnDiskWithDir("test", td, time.Second)
	assert.NoError(t, err)
	assert.NoError(t, odc.Set("foo", []string{"bar"}))
	res, err := odc.Get("foo")
	assert.NoError(t, err)
	assert.Equal(t, []string{"bar"}, res)

	time.Sleep(time.Second + 100*time.Millisecond)
	res, err = odc.Get("foo")
	assert.Error(t, err)
	assert.NotEqual(t, []string{"bar"}, res)
}
