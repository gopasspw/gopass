package cache

import (
	"os"
	"path/filepath"
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

func TestOnDiskNoExpiry(t *testing.T) {
	t.Parallel()

	for name, ttl := range map[string]time.Duration{"success_zero_ttl": 0, "success_negative_ttl": -time.Second} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			td := t.TempDir()
			odc, err := NewOnDiskWithDir("test", td, ttl)
			require.NoError(t, err)
			require.NoError(t, odc.Set("foo", []string{"bar"}))
			old := time.Now().Add(-365 * 24 * time.Hour)
			require.NoError(t, os.Chtimes(filepath.Join(td, "foo"), old, old))

			// Act
			res, err := odc.Get("foo")

			// Assert
			require.NoError(t, err)
			assert.Equal(t, []string{"bar"}, res)
		})
	}
}
