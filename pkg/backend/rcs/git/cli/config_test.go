package cli

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/justwatchcom/gopass/pkg/ctxutil"
	"github.com/justwatchcom/gopass/pkg/out"
	"github.com/stretchr/testify/assert"
)

func TestGitConfig(t *testing.T) {
	td, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()

	gitdir := filepath.Join(td, "git")
	assert.NoError(t, os.Mkdir(gitdir, 0755))

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithDebug(ctx, true)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	git, err := Init(ctx, gitdir, "Dead Beef", "dead.beef@example.org")
	assert.NoError(t, err)
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
