package cryptfs

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"filippo.io/age"
	"github.com/gopasspw/gopass/internal/backend"
	_ "github.com/gopasspw/gopass/internal/backend/storage/fs"
	_ "github.com/gopasspw/gopass/internal/backend/storage/gitfs"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestCryptFS(ctx context.Context, t *testing.T) (*Crypt, string) {
	td := t.TempDir()

	ctx = ctxutil.WithPasswordCallback(ctx, func(prompt string, _ bool) ([]byte, error) {
		return []byte(""), nil
	})

	// setup age identity
	xdgConfigHome := filepath.Join(td, ".config")
	t.Setenv("XDG_CONFIG_HOME", xdgConfigHome)
	identityPath := filepath.Join(xdgConfigHome, "gopass", "age", "identities")
	require.NoError(t, os.MkdirAll(filepath.Dir(identityPath), 0755))
	key, err := age.GenerateX25519Identity()
	require.NoError(t, err)
	err = os.WriteFile(identityPath, []byte(key.String()), 0600)
	require.NoError(t, err)

	// setup store
	storePath := filepath.Join(td, ".store")
	require.NoError(t, os.MkdirAll(storePath, 0755))

	// use fs backend for simplicity
	sub, err := backend.InitStorage(ctx, backend.FS, storePath)
	require.NoError(t, err)

	// create .age-recipients file
	recp := key.Recipient().String()
	err = os.WriteFile(filepath.Join(storePath, ".age-recipients"), []byte(recp), 0644)
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
	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	crypt, _ := newTestCryptFS(ctx, t)

	secret := []byte("my secret")
	name := "foo/bar"

	err := crypt.Set(ctx, name, secret)
	assert.NoError(t, err)

	ret, err := crypt.Get(ctx, name)
	assert.NoError(t, err)
	assert.Equal(t, secret, ret)
}

func TestList(t *testing.T) {
	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	crypt, _ := newTestCryptFS(ctx, t)

	err := crypt.Set(ctx, "foo/bar", []byte("1"))
	require.NoError(t, err)
	err = crypt.Set(ctx, "foo/baz", []byte("2"))
	require.NoError(t, err)
	err = crypt.Set(ctx, "qux/quux", []byte("3"))
	require.NoError(t, err)

	list, err := crypt.List(ctx, "")
	assert.NoError(t, err)
	assert.Equal(t, []string{"foo/", "qux/"}, list)

	list, err = crypt.List(ctx, "foo")
	assert.NoError(t, err)
	assert.Equal(t, []string{"bar", "baz"}, list)
}

func TestDelete(t *testing.T) {
	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	crypt, _ := newTestCryptFS(ctx, t)

	err := crypt.Set(ctx, "foo/bar", []byte("1"))
	require.NoError(t, err)
	assert.True(t, crypt.Exists(ctx, "foo/bar"))

	err = crypt.Delete(ctx, "foo/bar")
	assert.NoError(t, err)
	assert.False(t, crypt.Exists(ctx, "foo/bar"))
}

func TestMove(t *testing.T) {
	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	crypt, _ := newTestCryptFS(ctx, t)

	err := crypt.Set(ctx, "foo/bar", []byte("1"))
	require.NoError(t, err)
	assert.True(t, crypt.Exists(ctx, "foo/bar"))

	err = crypt.Move(ctx, "foo/bar", "foo/baz", true)
	assert.NoError(t, err)
	assert.False(t, crypt.Exists(ctx, "foo/bar"))
	assert.True(t, crypt.Exists(ctx, "foo/baz"))
}

func TestIsDir(t *testing.T) {
	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	crypt, _ := newTestCryptFS(ctx, t)

	err := crypt.Set(ctx, "foo/bar", []byte("1"))
	require.NoError(t, err)

	assert.True(t, crypt.IsDir(ctx, "foo"))
	assert.False(t, crypt.IsDir(ctx, "foo/bar"))
}

func TestGit(t *testing.T) {
	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithGitInit(ctx, true)
	ctx = ctxutil.WithUsername(ctx, "test")
	ctx = ctxutil.WithEmail(ctx, "test@example.com")
	ctx = ctxutil.WithPasswordCallback(ctx, func(prompt string, _ bool) ([]byte, error) {
		return []byte(""), nil
	})

	td := t.TempDir()

	// setup age identity
	xdgConfigHome := filepath.Join(td, ".config")
	t.Setenv("XDG_CONFIG_HOME", xdgConfigHome)
	identityPath := filepath.Join(xdgConfigHome, "gopass", "age", "identities")
	require.NoError(t, os.MkdirAll(filepath.Dir(identityPath), 0755))
	key, err := age.GenerateX25519Identity()
	require.NoError(t, err)
	err = os.WriteFile(identityPath, []byte(key.String()), 0600)
	require.NoError(t, err)

	// setup store
	storePath := filepath.Join(td, ".store")
	require.NoError(t, os.MkdirAll(storePath, 0755))

	// use gitfs backend
	sub, err := backend.InitStorage(ctx, backend.GitFS, storePath)
	require.NoError(t, err)

	// create .age-recipients file
	recp := key.Recipient().String()
	err = os.WriteFile(filepath.Join(storePath, ".age-recipients"), []byte(recp), 0644)
	require.NoError(t, err)

	// create cryptfs
	crypt, err := newCrypt(ctx, sub)
	require.NoError(t, err)
	err = crypt.saveMappings(ctx)
	require.NoError(t, err)

	// Set a secret
	err = crypt.Set(ctx, "foo/bar", []byte("1"))
	require.NoError(t, err)

	// Add and commit
	err = crypt.Add(ctx, crypt.Path())
	require.NoError(t, err)
	err = crypt.Commit(ctx, "initial commit")
	require.NoError(t, err)

	// Check revisions
	revs, err := crypt.Revisions(ctx, "foo/bar")
	assert.NoError(t, err)
	assert.Len(t, revs, 1)
}
