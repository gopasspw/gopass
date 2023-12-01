package backend

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectCrypto(t *testing.T) {
	for _, tc := range []struct {
		name string
		file string
	}{
		{
			name: "plain",
			file: ".plain-id",
		},
		{
			name: "gpg",
			file: ".gpg-id",
		},
		{
			name: "age",
			file: ".age-recipients",
		},
	} {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			ctx := config.NewContextInMemory()

			fsDir := filepath.Join(t.TempDir(), "fs")
			_ = os.RemoveAll(fsDir)
			require.NoError(t, os.MkdirAll(fsDir, 0o700))
			require.NoError(t, os.WriteFile(filepath.Join(fsDir, tc.file), []byte("foo"), 0o600))

			r, err := DetectStorage(ctx, fsDir)
			require.NoError(t, err)
			assert.NotNil(t, r)
			assert.Equal(t, "fs", r.Name())

			c, err := DetectCrypto(ctx, r)
			require.NoError(t, err, tc.name)
			require.NotNil(t, c, tc.name)
			assert.Equal(t, tc.name, c.Name())
		})
	}
}
