// +build !windows

package updater

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/gopasspw/gopass/pkg/ctxutil"
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
		err = updateCheckHost(u)
		if tc.ok {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestDownload(t *testing.T) {
	td, err := ioutil.TempDir("", "gopass-")
	require.NoError(t, err)

	defer func() {
		_ = os.RemoveAll(td)
	}()

	ctx := context.Background()
	ctx = ctxutil.WithHidden(ctx, true)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gzw := gzip.NewWriter(w)
		defer gzw.Close()
		tw := tar.NewWriter(gzw)
		defer tw.Close()
		tw.WriteHeader(&tar.Header{
			Name: "gopass",
			Mode: 0600,
			Size: 300,
		})
		for i := 0; i < 100; i++ {
			fmt.Fprintf(tw, "foo")
		}
	}))
	defer ts.Close()
	arc := filepath.Join(td, "gopass.tar.gz")
	assert.NoError(t, tryDownload(ctx, arc, ts.URL))
	dest := filepath.Join(td, "gopass")
	assert.NoError(t, extract(arc, dest))
}
