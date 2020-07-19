package gitfs

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitConfig(t *testing.T) {
	td, err := ioutil.TempDir("", "gopass-")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()

	gitdir := filepath.Join(td, "git")
	require.NoError(t, os.Mkdir(gitdir, 0755))

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	git, err := Init(ctx, gitdir, "Dead Beef", "dead.beef@example.org")
	require.NoError(t, err)
	un, err := git.ConfigGet(ctx, "user.name")
	assert.NoError(t, err)
	assert.Equal(t, "Dead Beef", un)

	assert.NoError(t, git.InitConfig(ctx, "Foo Bar", "foo.bar@example.org"))
	un, err = git.ConfigGet(ctx, "user.name")
	assert.NoError(t, err)
	assert.Equal(t, "Foo Bar", un)

	assert.NoError(t, git.ConfigSet(ctx, "user.name", "foo"))
	un, err = git.ConfigGet(ctx, "user.name")
	assert.NoError(t, err)
	assert.Equal(t, "foo", un)
}
