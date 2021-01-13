package backend

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/gopasspw/gopass/internal/gptest"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectStorage(t *testing.T) {
	ctx := context.Background()

	uv := gptest.UnsetVars("GOPASS_HOMEDIR")
	defer uv()

	td, err := ioutil.TempDir("", "gopass-")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()

	// all tests involving ondisk/age should set GOPASS_HOMEDIR
	os.Setenv("GOPASS_HOMEDIR", td)
	ctx = ctxutil.WithPasswordCallback(ctx, func(_ string) ([]byte, error) {
		debug.Log("static test password callback")
		return []byte("gopass"), nil
	})

	fsDir := filepath.Join(td, "fs")
	assert.NoError(t, os.MkdirAll(fsDir, 0700))

	ondiskDir := filepath.Join(td, "ondisk")
	assert.NoError(t, os.MkdirAll(ondiskDir, 0700))
	assert.NoError(t, ioutil.WriteFile(filepath.Join(ondiskDir, "index.gp1"), []byte("null"), 0600))

	t.Run("detect fs", func(t *testing.T) {
		r, err := DetectStorage(ctx, fsDir)
		assert.NoError(t, err)
		assert.NotNil(t, r)
		assert.Equal(t, "fs", r.Name())
	})

	t.Run("detect ondisk", func(t *testing.T) {
		r, err := DetectStorage(ctx, ondiskDir)
		// the "fake" index can't be decoded, so it must fail
		assert.Error(t, err)
		assert.Nil(t, r)
		assert.Equal(t, "ondisk", r.Name())
	})
}
