package fossilfs

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
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
	path := "/tmp/testpath"

	err := l.Handles(ctx, path)
	assert.Error(t, err)

	// Create a mock marker file for testing
	fsutil.MkdirAll(path, 0o700)
	defer fsutil.RemoveAll(path)
	marker := filepath.Join(path, CheckoutMarker)
	fsutil.WriteFile(marker, []byte("marker"), 0o600)

	err = l.Handles(ctx, path)
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
