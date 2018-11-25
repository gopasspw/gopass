package sub

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/gopasspw/gopass/pkg/backend"
	"github.com/gopasspw/gopass/pkg/store/secret"

	_ "github.com/gopasspw/gopass/pkg/backend/crypto"
	_ "github.com/gopasspw/gopass/pkg/backend/rcs"
	_ "github.com/gopasspw/gopass/pkg/backend/storage"

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
	ctx = backend.WithRCSBackendString(ctx, "noop")
	return New(
		ctx,
		&fakeConfig{},
		"",
		backend.FromPath(sd),
		sd,
		nil,
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
		filename := filepath.Join(dir, file+".gpg")
		if err := os.MkdirAll(filepath.Dir(filename), 0700); err != nil {
			return recipients, entries, err
		}
		if err := ioutil.WriteFile(filename, []byte{}, 0644); err != nil {
			return recipients, entries, err
		}
	}
	err := ioutil.WriteFile(filepath.Join(dir, ".gpg-id"), []byte(strings.Join(recipients, "\n")), 0600)
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
	require.NoError(t, s.Set(ctx, secName, secret.New("foo", "bar")))
	require.NoError(t, ioutil.WriteFile(filepath.Join(tempdir, "sub", "a", ".gpg-id"), []byte("foobar"), 0600))
	assert.Equal(t, filepath.Join("a", ".gpg-id"), s.idFile(ctx, secName))
	assert.Equal(t, true, s.Exists(ctx, secName))

	// test abort condition
	secName = "a"
	for i := 0; i < 100; i++ {
		secName += "/a"
	}
	require.NoError(t, s.Set(ctx, secName, secret.New("foo", "bar")))
	require.NoError(t, ioutil.WriteFile(filepath.Join(tempdir, "sub", "a", ".gpg-id"), []byte("foobar"), 0600))
	assert.Equal(t, ".gpg-id", s.idFile(ctx, secName))
}

func TestNew(t *testing.T) {
	ctx := context.Background()

	tempdir, err := ioutil.TempDir("", "gopass-")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	for _, tc := range []struct {
		ctx context.Context
		ok  bool
	}{
		{
			ctx: backend.WithStorageBackend(ctx, backend.InMem),
			ok:  true,
		},
		{
			ctx: backend.WithStorageBackend(ctx, -1),
			ok:  false,
		},
		{
			ctx: backend.WithRCSBackend(ctx, backend.GoGit),
			ok:  true,
		},
		{
			ctx: backend.WithRCSBackend(ctx, backend.GitCLI),
			ok:  true,
		},
		{
			ctx: backend.WithRCSBackend(ctx, backend.Noop),
			ok:  true,
		},
		{
			ctx: backend.WithRCSBackend(ctx, -1),
			ok:  false,
		},
		{
			ctx: backend.WithCryptoBackend(ctx, backend.GPGCLI),
			ok:  true,
		},
		{
			ctx: backend.WithCryptoBackend(ctx, backend.XC),
			ok:  true,
		},
		{
			ctx: backend.WithCryptoBackend(ctx, backend.Plain),
			ok:  true,
		},
		{
			ctx: backend.WithCryptoBackend(ctx, -1),
			ok:  false,
		},
	} {
		s, err := New(tc.ctx, nil, "", backend.FromPath(tempdir), tempdir, nil)
		if tc.ok {
			assert.NoError(t, err)
			assert.NotNil(t, s)
		} else {
			assert.Error(t, err)
		}
	}
}
