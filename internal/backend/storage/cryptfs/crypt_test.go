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

var password = "hunter2"

func newTestCryptFS(ctx context.Context, t *testing.T, td string) (*Crypt, string) {
	t.Helper()

	// setup age identity
	a, err := age.New(ctx, false, "")
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
	ctx = ctxutil.WithAgePassphrase(ctx, password)

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
	ctx = ctxutil.WithAgePassphrase(ctx, password)

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
	ctx = ctxutil.WithAgePassphrase(ctx, password)

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
	ctx = ctxutil.WithAgePassphrase(ctx, password)

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
	ctx = ctxutil.WithAgePassphrase(ctx, password)

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
	ctx = ctxutil.WithAgePassphrase(ctx, password)

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

	// A retry stages both the ciphertext and its encrypted name mapping.
	require.NoError(t, crypt.Set(ctx, "foo/baz", []byte("2")))
	staged, err := crypt.StageUncommitted(ctx, "foo/baz")
	require.NoError(t, err)
	assert.True(t, staged)
	require.NoError(t, crypt.Commit(ctx, "commit recovered file"))
	want, err := crypt.sub.Get(ctx, mappingFile)
	require.NoError(t, err)
	got, err := crypt.sub.GetRevision(ctx, mappingFile, "HEAD")
	require.NoError(t, err)
	assert.Equal(t, want, got)
	staged, err = crypt.StageUncommitted(ctx, "foo/baz")
	require.NoError(t, err)
	assert.False(t, staged)

	// The ciphertext can be committed before the mapping. Don't silently
	// commit the whole mapping when it also includes a name outside the batch.
	require.NoError(t, crypt.Set(ctx, "foo/first", []byte("3")))
	require.NoError(t, crypt.Set(ctx, "foo/second", []byte("4")))
	require.NoError(t, crypt.sub.Add(ctx, crypt.mappings["foo/first"]))
	require.NoError(t, crypt.sub.Commit(ctx, "only first ciphertext"))
	staged, err = crypt.StageUncommitted(ctx, "foo/first")
	require.ErrorContains(t, err, "unrelated pending changes")
	assert.False(t, staged)
	status, err := crypt.sub.Status(ctx)
	require.NoError(t, err)
	assert.NotContains(t, string(status), "Changes to be committed")

	staged, err = crypt.StageUncommitted(ctx, "foo/first", "foo/second")
	require.NoError(t, err)
	assert.True(t, staged)
	require.NoError(t, crypt.Commit(ctx, "recover both names"))
	want, err = crypt.sub.Get(ctx, mappingFile)
	require.NoError(t, err)
	got, err = crypt.sub.GetRevision(ctx, mappingFile, "HEAD")
	require.NoError(t, err)
	assert.Equal(t, want, got)
	_, err = crypt.GetRevision(ctx, "foo/second", "HEAD")
	require.NoError(t, err)

	// If both ciphertexts are already committed, the mapping alone must
	// still be staged even though neither ciphertext needs an update.
	require.NoError(t, crypt.Set(ctx, "foo/mapping-only", []byte("5")))
	require.NoError(t, crypt.sub.Add(ctx, crypt.mappings["foo/mapping-only"]))
	require.NoError(t, crypt.sub.Commit(ctx, "commit ciphertext before mapping"))
	staged, err = crypt.StageUncommitted(ctx, "foo/mapping-only")
	require.NoError(t, err)
	assert.True(t, staged)
	require.NoError(t, crypt.Commit(ctx, "recover only the mapping"))
	staged, err = crypt.StageUncommitted(ctx, "foo/mapping-only")
	require.NoError(t, err)
	assert.False(t, staged)
}
