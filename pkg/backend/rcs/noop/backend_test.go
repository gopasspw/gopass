package noop

import (
	"context"
	"testing"

	"github.com/blang/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoop(t *testing.T) {
	ctx := context.Background()

	g := New()
	assert.NoError(t, g.Add(ctx, "foo", "bar"))
	assert.NoError(t, g.Commit(ctx, "foobar"))
	assert.NoError(t, g.Push(ctx, "foo", "bar"))
	assert.NoError(t, g.Pull(ctx, "foo", "bar"))
	assert.NoError(t, g.Cmd(ctx, "foo", "bar"))
	assert.NoError(t, g.Init(ctx, "foo", "bar"))
	assert.NoError(t, g.InitConfig(ctx, "foo", "bar"))
	assert.Equal(t, g.Version(ctx), semver.Version{})
	assert.Equal(t, "noop", g.Name())
	assert.NoError(t, g.AddRemote(ctx, "foo", "bar"))
	_, err := g.Revisions(ctx, "foo")
	assert.Error(t, err)
	body, err := g.GetRevision(ctx, "foo", "bar")
	require.NoError(t, err)
	assert.Equal(t, "", string(body))
	assert.NoError(t, g.RemoveRemote(ctx, "foo"))
}
