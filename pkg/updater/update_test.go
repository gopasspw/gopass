package updater

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsUpdateable(t *testing.T) {
	ctx := context.Background()
	oldExec := executable
	defer func() {
		executable = oldExec
	}()

	td, err := ioutil.TempDir("", "gopass-")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()

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
				return "", fmt.Errorf("failed")
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
				return ioutil.WriteFile(filepath.Join(td, "gopass"), []byte("foobar"), 0555)
			},
			exec: func(context.Context) (string, error) {
				return filepath.Join(td, "gopass"), nil
			},
		},
		{
			name: "no write access to dir",
			pre: func() error {
				dir := filepath.Join(td, "bin")
				return os.Mkdir(dir, 0555)
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

func TestCheckHost(t *testing.T) {
	ctx := context.Background()

	for _, tc := range []struct {
		in string
		ok bool
	}{
		{
			in: "https://github.com/gopasspw/gopass/releases/download/v1.8.3/gopass-1.8.3-linux-amd64.tar.gz",
			ok: true,
		},
		{
			in: "http://localhost:8080/foo/bar.tar.gz",
			ok: true,
		},
	} {
		u, err := url.Parse(tc.in)
		require.NoError(t, err)
		err = updateCheckHost(ctx, u)
		if tc.ok {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}
