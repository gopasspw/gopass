package backend

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectStorage(t *testing.T) { //nolint:paralleltest
	ctx := context.Background()

	td, err := os.MkdirTemp("", "gopass-")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()

	// all tests involving age should set GOPASS_HOMEDIR
	t.Setenv("GOPASS_HOMEDIR", td)
	ctx = ctxutil.WithPasswordCallback(ctx, func(_ string, _ bool) ([]byte, error) {
		debug.Log("static test password callback")

		return []byte("gopass"), nil
	})

	fsDir := filepath.Join(td, "fs")
	assert.NoError(t, os.MkdirAll(fsDir, 0o700))

	t.Run("detect fs", func(t *testing.T) {
		r, err := DetectStorage(ctx, fsDir)
		assert.NoError(t, err)
		assert.NotNil(t, r)
		assert.Equal(t, "fs", r.Name())
	})
}
