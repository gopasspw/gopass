package cryptfs

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/backend/crypto/age"
	_ "github.com/gopasspw/gopass/internal/backend/storage/fs"
	_ "github.com/gopasspw/gopass/internal/backend/storage/gitfs"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	password = "hunter2"
)

func newTestCryptFS(ctx context.Context, t *testing.T, td string) (*Crypt, string) {
	t.Helper()

	// setup age identity
	a, err := age.New(ctx, "")
	require.NoError(t, err)

	recp, err := a.GenerateIdentity(ctx, "", "", password)
	require.NoError(t, err)

	// setup store
	storePath := filepath.Join(td, ".store")
	require.NoError(t, os.MkdirAll(storePath, 0o755))

	// use fs backend for simplicity
	sub, err := backend.InitStorage(ctx, backend.GitFS, storePath)
	require.NoError(t, err)

	// create .age-recipients file
	err = os.WriteFile(filepath.Join(storePath, ".age-recipients"), []byte(recp), 0o644)
	require.NoError(t, err)

	// create cryptfs
	crypt, err := newCrypt(ctx, sub)
	require.NoError(t, err)
	// save empty mapping
	err = crypt.saveMappings(ctx)
	require.NoError(t, err)

	return crypt, storePath
}

func TestSetGet(t *testing.T) {
	ctx := t.Context()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithPasswordCallback(ctx, func(prompt string, _ bool) ([]byte, error) {
		return []byte(password), nil
	})

	td := t.TempDir()
	t.Setenv("GOPASS_HOMEDIR", td)

	crypt, _ := newTestCryptFS(ctx, t, td)

	secret := []byte("my secret")
	name := "foo/bar"

	err := crypt.Set(ctx, name, secret)
	require.NoError(t, err)

	ret, err := crypt.Get(ctx, name)
	require.NoError(t, err)
	assert.Equal(t, secret, ret)
}

func TestList(t *testing.T) {
	ctx := t.Context()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithPasswordCallback(ctx, func(prompt string, _ bool) ([]byte, error) {
		return []byte(password), nil
	})

	td := t.TempDir()
	t.Setenv("GOPASS_HOMEDIR", td)

	crypt, _ := newTestCryptFS(ctx, t, td)

	err := crypt.Set(ctx, "foo/bar", []byte("1"))
	require.NoError(t, err)
	err = crypt.Set(ctx, "foo/baz", []byte("2"))
	require.NoError(t, err)
	err = crypt.Set(ctx, "qux/quux", []byte("3"))
	require.NoError(t, err)

	list, err := crypt.List(ctx, "")
	require.NoError(t, err)
	assert.Equal(t, []string{"foo/bar", "foo/baz", "qux/quux"}, list)

	list, err = crypt.List(ctx, "foo")
	require.NoError(t, err)
	assert.Equal(t, []string{"foo/bar", "foo/baz"}, list)
}

func TestDelete(t *testing.T) {
	ctx := t.Context()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithPasswordCallback(ctx, func(prompt string, _ bool) ([]byte, error) {
		return []byte(password), nil
	})

	td := t.TempDir()
	t.Setenv("GOPASS_HOMEDIR", td)

	crypt, _ := newTestCryptFS(ctx, t, td)

	err := crypt.Set(ctx, "foo/bar", []byte("1"))
	require.NoError(t, err)
	assert.True(t, crypt.Exists(ctx, "foo/bar"))

	err = crypt.Delete(ctx, "foo/bar")
	require.NoError(t, err)
	assert.False(t, crypt.Exists(ctx, "foo/bar"))
}

func TestMove(t *testing.T) {
	ctx := t.Context()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithPasswordCallback(ctx, func(prompt string, _ bool) ([]byte, error) {
		return []byte(password), nil
	})

	td := t.TempDir()
	t.Setenv("GOPASS_HOMEDIR", td)

	crypt, _ := newTestCryptFS(ctx, t, td)

	err := crypt.Set(ctx, "foo/bar", []byte("1"))
	require.NoError(t, err)
	assert.True(t, crypt.Exists(ctx, "foo/bar"))

	err = crypt.Move(ctx, "foo/bar", "foo/baz", true)
	require.NoError(t, err)
	assert.False(t, crypt.Exists(ctx, "foo/bar"))
	assert.True(t, crypt.Exists(ctx, "foo/baz"))
}

func TestIsDir(t *testing.T) {
	ctx := t.Context()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithPasswordCallback(ctx, func(prompt string, _ bool) ([]byte, error) {
		return []byte(password), nil
	})

	td := t.TempDir()
	t.Setenv("GOPASS_HOMEDIR", td)

	crypt, _ := newTestCryptFS(ctx, t, td)

	err := crypt.Set(ctx, "foo/bar", []byte("1"))
	require.NoError(t, err)

	assert.True(t, crypt.IsDir(ctx, "foo"))
	assert.False(t, crypt.IsDir(ctx, "foo/bar"))
}

func TestGit(t *testing.T) {
	ctx := t.Context()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithGitInit(ctx, true)
	ctx = ctxutil.WithUsername(ctx, "test")
	ctx = ctxutil.WithEmail(ctx, "test@example.com")
	ctx = ctxutil.WithPasswordCallback(ctx, func(prompt string, _ bool) ([]byte, error) {
		return []byte(password), nil
	})

	td := t.TempDir()
	t.Setenv("GOPASS_HOMEDIR", td)

	crypt, _ := newTestCryptFS(ctx, t, td)

	// Set a secret
	err := crypt.Set(ctx, "foo/bar", []byte("1"))
	require.NoError(t, err)

	// Add and commit
	err = crypt.Add(ctx, crypt.Path())
	require.NoError(t, err)
	err = crypt.Commit(ctx, "initial commit")
	require.NoError(t, err)

	// Check revisions
	revs, err := crypt.Revisions(ctx, "foo/bar")
	require.NoError(t, err)
	assert.Len(t, revs, 1)
}
