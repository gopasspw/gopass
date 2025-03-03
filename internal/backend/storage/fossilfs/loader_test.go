package fossilfs

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoader_New(t *testing.T) {
	l := loader{}
	ctx := context.Background()
	path := "/tmp/testpath"

	storage, err := l.New(ctx, path)
	assert.NoError(t, err)
	assert.NotNil(t, storage)
}

func TestLoader_Open(t *testing.T) {
	l := loader{}
	ctx := context.Background()
	path := "/tmp/testpath"

	storage, err := l.Open(ctx, path)
	assert.NoError(t, err)
	assert.NotNil(t, storage)
}

func TestLoader_Clone(t *testing.T) {
	l := loader{}
	ctx := context.Background()
	repo := "https://example.com/repo.git"
	path := "/tmp/testpath"

	storage, err := l.Clone(ctx, repo, path)
	assert.NoError(t, err)
	assert.NotNil(t, storage)
}

func TestLoader_Init(t *testing.T) {
	l := loader{}
	ctx := context.Background()
	path := "/tmp/testpath"

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

	// Create a mock marker file for testing
	require.NoError(t, os.MkdirAll(td, 0o700))
	marker := filepath.Join(td, CheckoutMarker)
	os.WriteFile(marker, []byte("marker"), 0o600)

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
