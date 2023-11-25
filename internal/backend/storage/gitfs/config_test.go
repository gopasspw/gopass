package gitfs

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitConfig(t *testing.T) {
	gitdir := filepath.Join(t.TempDir(), "git")
	require.NoError(t, os.Mkdir(gitdir, 0o755))

	ctx := config.NewNoWrites().WithConfig(context.Background())
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	git, err := Init(ctx, gitdir, "Dead Beef", "dead.beef@example.org")
	require.NoError(t, err)
	un, err := git.ConfigGet(ctx, "user.name")
	require.NoError(t, err)
	assert.Equal(t, "Dead Beef", un)

	require.NoError(t, git.InitConfig(ctx, "Foo Bar", "foo.bar@example.org"))
	un, err = git.ConfigGet(ctx, "user.name")
	require.NoError(t, err)
	assert.Equal(t, "Foo Bar", un)

	require.NoError(t, git.ConfigSet(ctx, "user.name", "foo"))
	un, err = git.ConfigGet(ctx, "user.name")
	require.NoError(t, err)
	assert.Equal(t, "foo", un)
}
