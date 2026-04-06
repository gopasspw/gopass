package backend

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectStorage(t *testing.T) {
	ctx := config.NewContextInMemory()

	td := t.TempDir()

	// all tests involving age should set GOPASS_HOMEDIR
	t.Setenv("GOPASS_HOMEDIR", td)
	ctx = ctxutil.WithAgePassphrase(ctx, "gopass")

	fsDir := filepath.Join(td, "fs")
	require.NoError(t, os.MkdirAll(fsDir, 0o700))

	t.Run("detect fs", func(t *testing.T) {
		r, err := DetectStorage(ctx, fsDir)
		require.NoError(t, err)
		assert.NotNil(t, r)
		assert.Equal(t, "fs", r.Name())
	})
}
