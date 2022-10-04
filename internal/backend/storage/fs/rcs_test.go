package fs

import (
	"context"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRCS(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	path := t.TempDir()

	g := New(path)
	// the fs backend does not support the RCS operations
	assert.Error(t, g.Add(ctx, "foo", "bar"))
	assert.Error(t, g.Commit(ctx, "foobar"))
	assert.Error(t, g.Push(ctx, "foo", "bar"))
	assert.Error(t, g.Pull(ctx, "foo", "bar"))
	assert.NoError(t, g.Cmd(ctx, "foo", "bar"))
	assert.Error(t, g.Init(ctx, "foo", "bar"))
	assert.NoError(t, g.InitConfig(ctx, "foo", "bar"))
	assert.Equal(t, true, g.Version(ctx).EQ(semver.Version{}), "Version eq 0.0.0")
	assert.Equal(t, "fs", g.Name())
	assert.Error(t, g.AddRemote(ctx, "foo", "bar"))
	revs, err := g.Revisions(ctx, "foo")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(revs))
	body, err := g.GetRevision(ctx, "foo", "latest")
	require.Error(t, err)
	assert.Equal(t, "", string(body))
	assert.Error(t, g.RemoveRemote(ctx, "foo"))
}
