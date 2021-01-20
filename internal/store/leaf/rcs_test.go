package leaf

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/pkg/ctxutil"

	"github.com/blang/semver/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGit(t *testing.T) {
	ctx := context.Background()

	tempdir, err := ioutil.TempDir("", "gopass-")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	s, err := createSubStore(tempdir)
	require.NoError(t, err)

	require.NotNil(t, s.Storage())
	require.Equal(t, "fs", s.Storage().Name())
	assert.NoError(t, s.Storage().InitConfig(ctx, "foo", "bar@baz.com"))
	assert.Equal(t, semver.Version{Minor: 1}, s.Storage().Version(ctx))
	assert.NoError(t, s.Storage().AddRemote(ctx, "foo", "bar"))
	assert.NoError(t, s.Storage().Pull(ctx, "origin", "master"))
	assert.NoError(t, s.Storage().Push(ctx, "origin", "master"))

	assert.NoError(t, s.GitInit(ctx))
	assert.NoError(t, s.GitInit(backend.WithStorageBackend(ctx, backend.FS)))
	assert.Error(t, s.GitInit(backend.WithStorageBackend(ctx, -1)))

	ctx = ctxutil.WithUsername(ctx, "foo")
	ctx = ctxutil.WithEmail(ctx, "foo@baz.com")
	assert.NoError(t, s.GitInit(backend.WithStorageBackend(ctx, backend.GitFS)))
}

func TestGitRevisions(t *testing.T) {
	ctx := context.Background()

	tempdir, err := ioutil.TempDir("", "gopass-")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	s, err := createSubStore(tempdir)
	require.NoError(t, err)

	require.NotNil(t, s.Storage())
	require.Equal(t, "fs", s.Storage().Name())
	assert.NoError(t, s.Storage().InitConfig(ctx, "foo", "bar@baz.com"))

	revs, err := s.ListRevisions(ctx, "foo")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(revs))

	sec, err := s.GetRevision(ctx, "foo", "bar")
	require.NoError(t, err)
	assert.Equal(t, "foo", sec.Password())
}
