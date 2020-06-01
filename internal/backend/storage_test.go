package backend

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectStorage(t *testing.T) {
	ctx := context.Background()

	td, err := ioutil.TempDir("", "gopass-")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()

	fsDir := filepath.Join(td, "fs")
	assert.NoError(t, os.MkdirAll(fsDir, 0700))

	inmemDir := "//gopass/inmem"

	ondiskDir := filepath.Join(td, "ondisk")
	assert.NoError(t, os.MkdirAll(ondiskDir, 0700))
	assert.NoError(t, ioutil.WriteFile(filepath.Join(ondiskDir, "index.pb"), []byte("null"), 0600))

	r, err := DetectStorage(ctx, fsDir)
	assert.NoError(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, "fs", r.Name())

	r, err = DetectStorage(ctx, inmemDir)
	assert.NoError(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, "inmem", r.Name())

	t.Skip("WIP")

	r, err = DetectStorage(ctx, ondiskDir)
	assert.NoError(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, "ondisk", r.Name())
}
