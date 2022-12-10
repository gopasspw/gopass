//go:build !windows
// +build !windows

package updater

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:wrapcheck
func TestIsUpdateable(t *testing.T) {
	ctx := context.Background()
	oldExec := executable

	defer func() {
		executable = oldExec
	}()

	td := t.TempDir()

	for _, tc := range []struct {
		name string
		pre  func() error
		exec func(context.Context) (string, error)
		post func() error
		ok   bool
	}{
		{
			name: "executable error",
			exec: func(context.Context) (string, error) {
				return "", fmt.Errorf("failed") //nolint:goerr113
			},
		},
		{
			name: "test binary",
			exec: func(context.Context) (string, error) {
				return "action.test", nil
			},
			ok: true,
		},
		{
			name: "force update",
			pre: func() error {
				return os.Setenv("GOPASS_FORCE_UPDATE", "true")
			},
			exec: func(context.Context) (string, error) {
				return "", nil
			},
			post: func() error {
				return os.Unsetenv("GOPASS_FORCE_UPDATE")
			},
			ok: true,
		},
		{
			name: "update in gopath",
			pre: func() error {
				return os.Setenv("GOPATH", "/tmp/foo")
			},
			exec: func(context.Context) (string, error) {
				return "/tmp/foo/gopass", nil
			},
		},
		{
			name: "stat error",
			exec: func(context.Context) (string, error) {
				return "/tmp/foo/gopass", nil
			},
		},
		{
			name: "no regular file",
			exec: func(context.Context) (string, error) {
				return td, nil
			},
		},
		{
			name: "no write access to file",
			pre: func() error {
				return os.WriteFile(filepath.Join(td, "gopass"), []byte("foobar"), 0o555)
			},
			exec: func(context.Context) (string, error) {
				return filepath.Join(td, "gopass"), nil
			},
		},
		{
			name: "no write access to dir",
			pre: func() error {
				dir := filepath.Join(td, "bin")

				return os.Mkdir(dir, 0o555)
			},
			exec: func(context.Context) (string, error) {
				return filepath.Join(td, "bin"), nil
			},
		},
	} {
		if tc.pre != nil {
			require.NoError(t, tc.pre(), tc.name)
		}

		executable = tc.exec

		err := IsUpdateable(ctx)
		if tc.ok {
			assert.NoError(t, err, tc.name)
		} else {
			assert.Error(t, err, tc.name)
		}

		if tc.post != nil {
			assert.NoError(t, tc.post(), tc.name)
		}
	}
}
