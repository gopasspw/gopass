package mock

import (
	"context"
	"testing"

	"github.com/blang/semver"
	"github.com/stretchr/testify/assert"
)

func TestGitMock(t *testing.T) {
	ctx := context.Background()

	g := New()
	assert.NoError(t, g.Add(ctx, "foo", "bar"))
	assert.NoError(t, g.Commit(ctx, "foobar"))
	assert.NoError(t, g.Push(ctx, "foo", "bar"))
	assert.NoError(t, g.Cmd(ctx, "foo", "bar"))
	assert.NoError(t, g.Init(ctx, "foo", "bar", "baz"))
	assert.NoError(t, g.InitConfig(ctx, "foo", "bar", "baz"))
	assert.Equal(t, g.Version(ctx), semver.Version{})
}
