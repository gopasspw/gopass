package fossilfs

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMarker(t *testing.T, path string) {
	// Create a mock marker file for testing
	require.NoError(t, os.MkdirAll(path, 0o700))
	marker := filepath.Join(path, CheckoutMarker)
	require.NoError(t, os.WriteFile(marker, []byte("marker"), 0o600))
}

func TestLoader_New(t *testing.T) {
	l := loader{}
	ctx := context.Background()
	path := t.TempDir()
	createMarker(t, path)

	storage, err := l.New(ctx, path)
	assert.NoError(t, err)
	assert.NotNil(t, storage)
}

func TestLoader_Open(t *testing.T) {
	l := loader{}
	ctx := context.Background()
	path := t.TempDir()
	createMarker(t, path)

	storage, err := l.Open(ctx, path)
	assert.NoError(t, err)
	assert.NotNil(t, storage)
}

func TestLoader_Clone(t *testing.T) {
	t.Skip("needs fossil binary and valid remote")

	l := loader{}
	ctx := context.Background()
	repo := "https://example.com/repo.git"
	path := t.TempDir()
	createMarker(t, path)

	storage, err := l.Clone(ctx, repo, path)
	assert.NoError(t, err)
	assert.NotNil(t, storage)
}

func TestLoader_Init(t *testing.T) {
	t.Skip("needs fossil binary")

	l := loader{}
	ctx := context.Background()
	path := t.TempDir()
	createMarker(t, path)

	storage, err := l.Init(ctx, path)
	assert.NoError(t, err)
	assert.NotNil(t, storage)
}

func TestLoader_Handles(t *testing.T) {
	l := loader{}
	ctx := context.Background()
	td := t.TempDir()

	err := l.Handles(ctx, td)
	assert.Error(t, err)

	createMarker(t, td)

	err = l.Handles(ctx, td)
	assert.NoError(t, err)
}

func TestLoader_Priority(t *testing.T) {
	l := loader{}
	assert.Equal(t, 2, l.Priority())
}

func TestLoader_String(t *testing.T) {
	l := loader{}
	assert.Equal(t, name, l.String())
}
