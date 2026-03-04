package fs

import (
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRCS(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()
	path := t.TempDir()

	g := New(path)
	// the fs backend does not support the RCS operations
	require.Error(t, g.Add(ctx, "foo", "bar"))
	require.Error(t, g.Commit(ctx, "foobar"))
	require.Error(t, g.Push(ctx, "foo", "bar"))
	require.Error(t, g.Pull(ctx, "foo", "bar"))
	require.NoError(t, g.Cmd(ctx, "foo", "bar"))
	require.Error(t, g.Init(ctx, "foo", "bar"))
	require.NoError(t, g.InitConfig(ctx, "foo", "bar"))
	assert.Equal(t, "fs", g.Name())
	require.Error(t, g.AddRemote(ctx, "foo", "bar"))
	revs, err := g.Revisions(ctx, "foo")
	require.Error(t, err)
	assert.Len(t, revs, 1)
	body, err := g.GetRevision(ctx, "foo", "latest")
	require.Error(t, err)
	assert.Empty(t, string(body))
	require.Error(t, g.RemoveRemote(ctx, "foo"))
}
