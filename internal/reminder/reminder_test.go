package reminder

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Setenv("GOPASS_HOMEDIR", t.TempDir())

	store, err := New()
	require.NoError(t, err)
	assert.NotNil(t, store)
}

func TestLastSeen(t *testing.T) {
	t.Setenv("GOPASS_HOMEDIR", t.TempDir())

	store, err := New()
	require.NoError(t, err)
	assert.NotNil(t, store)

	key := "test-key"
	now := time.Now().Format(time.RFC3339)
	err = store.cache.Set(key, []string{now})
	require.NoError(t, err)

	lastSeen := store.LastSeen(key)
	assert.Equal(t, now, lastSeen.Format(time.RFC3339))
}

func TestReset(t *testing.T) {
	t.Setenv("GOPASS_HOMEDIR", t.TempDir())

	store, err := New()
	require.NoError(t, err)
	assert.NotNil(t, store)

	key := "test-key"
	err = store.Reset(key)
	require.NoError(t, err)

	lastSeen := store.LastSeen(key)
	assert.WithinDuration(t, time.Now(), lastSeen, time.Second)
}

func TestOverdue(t *testing.T) {
	t.Setenv("GOPASS_HOMEDIR", t.TempDir())

	store, err := New()
	require.NoError(t, err)
	assert.NotNil(t, store)

	key := "test-key"
	err = store.Reset(key)
	require.NoError(t, err)

	overdue := store.Overdue(key)
	assert.False(t, overdue)

	// Simulate overdue by setting the last seen time to more than 90 days ago
	past := time.Now().Add(-91 * 24 * time.Hour).Format(time.RFC3339)
	require.NoError(t, store.cache.Set(key, []string{past}))
	require.NoError(t, err)
	require.NoError(t, store.cache.Set("overdue", []string{time.Now().Add(-25 * time.Hour).Format(time.RFC3339)}))

	t.Logf("last seen: %s, %s ago", store.LastSeen(key), time.Since(store.LastSeen(key)))
	overdue = store.Overdue(key)
	assert.True(t, overdue)
}
