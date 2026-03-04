package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOnDisk(t *testing.T) {
	t.Parallel()

	td := t.TempDir()

	odc, err := NewOnDiskWithDir("test", td, time.Hour)
	require.NoError(t, err)

	require.NoError(t, odc.Set("foo", []string{"bar"}))
	res, err := odc.Get("foo")
	require.NoError(t, err)
	assert.Equal(t, []string{"bar"}, res)

	require.Error(t, odc.Remove("bar"))
	require.NoError(t, odc.Remove("foo"))
	require.NoError(t, odc.Purge())
}

func TestOnDiskExpiry(t *testing.T) {
	t.Parallel()

	td := t.TempDir()

	odc, err := NewOnDiskWithDir("test", td, time.Second)
	require.NoError(t, err)
	require.NoError(t, odc.Set("foo", []string{"bar"}))
	res, err := odc.Get("foo")
	require.NoError(t, err)
	assert.Equal(t, []string{"bar"}, res)

	time.Sleep(time.Second + 100*time.Millisecond)
	res, err = odc.Get("foo")
	require.Error(t, err)
	assert.NotEqual(t, []string{"bar"}, res)
}
