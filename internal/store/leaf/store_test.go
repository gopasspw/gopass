package leaf

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"

	_ "github.com/gopasspw/gopass/internal/backend/crypto"
	"github.com/gopasspw/gopass/internal/backend/crypto/plain"
	_ "github.com/gopasspw/gopass/internal/backend/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createSubStore(dir string) (*Store, error) {
	sd := filepath.Join(dir, "sub")
	_, _, err := createStore(sd, nil, nil)
	if err != nil {
		return nil, err
	}

	if err := os.Setenv("GOPASS_CONFIG", filepath.Join(dir, ".gopass.yml")); err != nil {
		return nil, err
	}
	if err := os.Setenv("GOPASS_HOMEDIR", dir); err != nil {
		return nil, err
	}
	if err := os.Unsetenv("PAGER"); err != nil {
		return nil, err
	}
	if err := os.Setenv("CHECKPOINT_DISABLE", "true"); err != nil {
		return nil, err
	}
	if err := os.Setenv("GOPASS_NO_NOTIFY", "true"); err != nil {
		return nil, err
	}
	if err := os.Setenv("GOPASS_DISABLE_ENCRYPTION", "true"); err != nil {
		return nil, err
	}

	gpgDir := filepath.Join(dir, ".gnupg")
	if err := os.Setenv("GNUPGHOME", gpgDir); err != nil {
		return nil, err
	}

	ctx := context.Background()
	ctx = backend.WithCryptoBackendString(ctx, "plain")
	ctx = backend.WithStorageBackendString(ctx, "fs")
	return New(
		ctx,
		"",
		sd,
	)
}

func createStore(dir string, recipients, entries []string) ([]string, []string, error) {
	if recipients == nil {
		recipients = []string{
			"0xDEADBEEF",
			"0xFEEDBEEF",
		}
	}
	if entries == nil {
		entries = []string{
			"foo/bar/baz",
			"baz/ing/a",
		}
	}
	sort.Strings(entries)
	for _, file := range entries {
		filename := filepath.Join(dir, file+"."+plain.Ext)
		if err := os.MkdirAll(filepath.Dir(filename), 0700); err != nil {
			return recipients, entries, err
		}
		if err := ioutil.WriteFile(filename, []byte{}, 0644); err != nil {
			return recipients, entries, err
		}
	}
	err := ioutil.WriteFile(filepath.Join(dir, plain.IDFile), []byte(strings.Join(recipients, "\n")), 0600)
	return recipients, entries, err
}

func TestStore(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "gopass-")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	s, err := createSubStore(tempdir)
	require.NoError(t, err)

	if !s.Equals(s) {
		t.Errorf("Should be equal to myself")
	}
}

func TestIdFile(t *testing.T) {
	ctx := context.Background()

	tempdir, err := ioutil.TempDir("", "gopass-")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	s, err := createSubStore(tempdir)
	require.NoError(t, err)

	// test sub-id
	secName := "a"
	for i := 0; i < 99; i++ {
		secName += "/a"
	}
	sec := &secrets.Plain{}
	sec.Set("foo", "bar")
	sec.WriteString("bar")
	require.NoError(t, s.Set(ctx, secName, sec))
	require.NoError(t, ioutil.WriteFile(filepath.Join(tempdir, "sub", "a", plain.IDFile), []byte("foobar"), 0600))
	assert.Equal(t, filepath.Join("a", plain.IDFile), s.idFile(ctx, secName))
	assert.True(t, s.Exists(ctx, secName))

	// test abort condition
	secName = "a"
	for i := 0; i < 100; i++ {
		secName += "/a"
	}
	require.NoError(t, s.Set(ctx, secName, sec))
	require.NoError(t, ioutil.WriteFile(filepath.Join(tempdir, "sub", "a", ".gpg-id"), []byte("foobar"), 0600))
	assert.Equal(t, plain.IDFile, s.idFile(ctx, secName))
}

func TestNew(t *testing.T) {
	ctx := context.Background()

	tempdir, err := ioutil.TempDir("", "gopass-")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	for _, tc := range []struct {
		dsc string
		dir string
		ctx context.Context
		ok  bool
	}{
		{
			dsc: "Invalid Storage",
			ctx: backend.WithStorageBackend(ctx, -1),
			// ok:  false, // TODO clarify
			ok: true,
		},
		{
			dsc: "GitFS Storage",
			dir: tempdir,
			ctx: backend.WithCryptoBackend(backend.WithStorageBackend(ctx, backend.GitFS), backend.Plain),
			ok:  true,
		},
		{
			dsc: "FS Storage",
			dir: tempdir,
			ctx: backend.WithCryptoBackend(backend.WithStorageBackend(ctx, backend.FS), backend.Plain),
			ok:  true,
		},
		{
			dsc: "GPG Crypto",
			dir: tempdir,
			ctx: backend.WithCryptoBackend(ctx, backend.GPGCLI),
			ok:  true,
		},
		{
			dsc: "Plain Crypto",
			dir: tempdir,
			ctx: backend.WithCryptoBackend(ctx, backend.Plain),
			ok:  true,
		},
		{
			dsc: "Invalid Crypto",
			dir: tempdir,
			ctx: backend.WithCryptoBackend(ctx, -1),
			// ok:  false, // TODO clarify
			ok: true,
		},
	} {
		s, err := New(tc.ctx, "", tempdir)
		if tc.ok {
			assert.NoError(t, err, tc.dsc)
			assert.NotNil(t, s, tc.dsc)
		} else {
			assert.Error(t, err, tc.dsc)
		}
	}
}
