package leaf

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/gopasspw/gopass/internal/backend"
	_ "github.com/gopasspw/gopass/internal/backend/crypto"
	"github.com/gopasspw/gopass/internal/backend/crypto/plain"
	_ "github.com/gopasspw/gopass/internal/backend/storage"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createSubStore(t *testing.T) (*Store, error) {
	t.Helper()

	dir := t.TempDir()
	sd := filepath.Join(dir, "sub")

	_, _, err := createStore(sd, nil, nil)
	if err != nil {
		return nil, err
	}

	t.Setenv("GOPASS_HOMEDIR", dir)
	t.Setenv("CHECKPOINT_DISABLE", "true")
	t.Setenv("GOPASS_NO_NOTIFY", "true")
	t.Setenv("GOPASS_DISABLE_ENCRYPTION", "true")
	t.Setenv("GNUPGHOME", filepath.Join(dir, ".gnupg"))

	if err := os.Unsetenv("PAGER"); err != nil {
		return nil, err
	}

	ctx := config.NewContextInMemory()
	ctx, err = backend.WithCryptoBackendString(ctx, "plain")
	require.NoError(t, err)
	ctx, err = backend.WithStorageBackendString(ctx, "fs")
	require.NoError(t, err)

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
		if err := os.MkdirAll(filepath.Dir(filename), 0o700); err != nil {
			return recipients, entries, err
		}

		if err := os.WriteFile(filename, []byte{}, 0o644); err != nil {
			return recipients, entries, err
		}
	}

	err := os.WriteFile(filepath.Join(dir, plain.IDFile), []byte(strings.Join(recipients, "\n")), 0o600)

	return recipients, entries, err
}

func TestStore(t *testing.T) {
	s, err := createSubStore(t)
	require.NoError(t, err)

	if !s.Equals(s) {
		t.Errorf("Should be equal to myself")
	}
}

func TestIdFile(t *testing.T) {
	ctx := config.NewContextInMemory()

	s, err := createSubStore(t)
	require.NoError(t, err)

	// test sub-id
	secName := "a"
	for range 99 {
		secName += "/a"
	}

	sec := secrets.NewAKV()

	_ = sec.Set("foo", "bar")
	_, err = sec.Write([]byte("bar"))
	require.NoError(t, err)
	require.NoError(t, s.Set(ctx, secName, sec))
	require.NoError(t, os.WriteFile(filepath.Join(s.path, "a", plain.IDFile), []byte("foobar"), 0o600))
	assert.Equal(t, filepath.Join("a", plain.IDFile), s.idFile(ctx, secName))
	assert.True(t, s.Exists(ctx, secName))

	// test abort condition
	secName = "a"
	for range 100 {
		secName += "/a"
	}
	require.NoError(t, s.Set(ctx, secName, sec))
	require.NoError(t, os.WriteFile(filepath.Join(s.path, "a", ".gpg-id"), []byte("foobar"), 0o600))
	assert.Equal(t, plain.IDFile, s.idFile(ctx, secName))
}

func TestNew(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		dsc   string
		noDir bool
		ctx   context.Context //nolint:containedctx
		ok    bool
	}{
		{
			dsc:   "Invalid Storage",
			ctx:   backend.WithStorageBackend(config.NewContextInMemory(), -1),
			noDir: true,
			ok:    false,
		},
		{
			dsc: "GitFS Storage",
			ctx: backend.WithCryptoBackend(backend.WithStorageBackend(config.NewContextInMemory(), backend.GitFS), backend.Plain),
			ok:  true,
		},
		{
			dsc: "FS Storage",
			ctx: backend.WithCryptoBackend(backend.WithStorageBackend(config.NewContextInMemory(), backend.FS), backend.Plain),
			ok:  true,
		},
		{
			dsc: "GPG Crypto",
			ctx: backend.WithCryptoBackend(config.NewContextInMemory(), backend.GPGCLI),
			ok:  true,
		},
		{
			dsc: "Plain Crypto",
			ctx: backend.WithCryptoBackend(config.NewContextInMemory(), backend.Plain),
			ok:  true,
		},
		{
			dsc: "Invalid Crypto",
			ctx: backend.WithCryptoBackend(config.NewContextInMemory(), -1),
			ok:  false,
		},
	} {
		t.Run(tc.dsc, func(t *testing.T) {
			t.Parallel()

			var tempdir string
			if !tc.noDir {
				tempdir = t.TempDir()
			}

			s, err := New(tc.ctx, "", tempdir)
			if !tc.ok {
				require.Error(t, err, tc.dsc)

				return
			}

			require.NoError(t, err, tc.dsc)
			assert.NotNil(t, s, tc.dsc)
		})
	}
}

// TestPassfileNormalization verifies that passfile() normalizes secret names
// when core.casefold is enabled, and is a no-op when it is disabled.
// On case-sensitive platforms (Linux) NormalizeSecretName is always a no-op,
// so we only verify the config-gating here.
func TestPassfileNormalization(t *testing.T) {
	// Cannot be parallel: createSubStore uses t.Setenv.

	s, err := createSubStore(t)
	require.NoError(t, err)

	// Build a context with casefold=false.
	cfgOff := config.NewInMemory()
	require.NoError(t, cfgOff.Set("", "core.casefold", "false"))
	ctxOff := cfgOff.WithConfig(context.Background())
	ctxOff, err = backend.WithCryptoBackendString(ctxOff, "plain")
	require.NoError(t, err)

	// With casefold=false the passfile path must equal the public Passfile path.
	assert.Equal(t, s.Passfile("foo/Bar"), s.passfile(ctxOff, "foo/Bar"),
		"passfile without casefold must equal Passfile()")

	// Build a context with casefold=true.
	cfgOn := config.NewInMemory()
	require.NoError(t, cfgOn.Set("", "core.casefold", "true"))
	ctxOn := cfgOn.WithConfig(context.Background())
	ctxOn, err = backend.WithCryptoBackendString(ctxOn, "plain")
	require.NoError(t, err)

	// With casefold=true the result must be deterministic.
	p1 := s.passfile(ctxOn, "foo/Bar")
	p2 := s.passfile(ctxOn, "foo/Bar")
	assert.Equal(t, p1, p2, "passfile must be deterministic")
	assert.NotEmpty(t, p1)
}
