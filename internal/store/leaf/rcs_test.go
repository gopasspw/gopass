package leaf

import (
	"testing"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGit(t *testing.T) {
	ctx := config.NewContextInMemory()

	s, err := createSubStore(t)
	require.NoError(t, err)

	require.NotNil(t, s.Storage())
	require.Equal(t, "fs", s.Storage().Name())
	require.NoError(t, s.Storage().InitConfig(ctx, "foo", "bar@baz.com"))
	// RCS ops not supported by the fs backend
	require.Error(t, s.Storage().AddRemote(ctx, "foo", "bar"))
	require.Error(t, s.Storage().Pull(ctx, "origin", "master"))
	require.Error(t, s.Storage().Push(ctx, "origin", "master"))

	require.NoError(t, s.GitInit(ctx))
	require.NoError(t, s.GitInit(backend.WithStorageBackend(ctx, backend.FS)))
	require.Error(t, s.GitInit(backend.WithStorageBackend(ctx, -1)))

	ctx = ctxutil.WithUsername(ctx, "foo")
	ctx = ctxutil.WithEmail(ctx, "foo@baz.com")
	require.NoError(t, s.GitInit(backend.WithStorageBackend(ctx, backend.GitFS)))
}

func TestGitRevisions(t *testing.T) {
	ctx := config.NewContextInMemory()

	s, err := createSubStore(t)
	require.NoError(t, err)

	require.NotNil(t, s.Storage())
	require.Equal(t, "fs", s.Storage().Name())
	require.NoError(t, s.Storage().InitConfig(ctx, "foo", "bar@baz.com"))

	revs, err := s.ListRevisions(ctx, "foo")
	require.Error(t, err)  // not supported by the fs backend
	assert.Len(t, revs, 1) // but it will still give a fake "latest" rev

	sec, err := s.GetRevision(ctx, "foo", "latest")
	require.Error(t, err)
	assert.Nil(t, sec)
}
