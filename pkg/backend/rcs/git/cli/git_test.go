package cli

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/out"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGit(t *testing.T) {
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
	require.NotNil(t, git)

	sv := git.Version(ctx)
	assert.NotEqual(t, "", sv.String())

	assert.Equal(t, true, git.IsInitialized())
	tf := filepath.Join(gitdir, "some-file")
	require.NoError(t, ioutil.WriteFile(tf, []byte("foobar"), 0644))
	assert.NoError(t, git.Add(ctx, "some-file"))
	assert.Equal(t, true, git.HasStagedChanges(ctx))
	assert.NoError(t, git.Commit(ctx, "added some-file"))
	assert.Equal(t, false, git.HasStagedChanges(ctx))

	assert.Error(t, git.Push(ctx, "origin", "master"))
	assert.Error(t, git.Pull(ctx, "origin", "master"))

	git, err = Open(gitdir, "")
	require.NoError(t, err)
	require.NotNil(t, git)
	assert.Equal(t, "git", git.Name())
	assert.NoError(t, git.AddRemote(ctx, "foo", "file:///tmp/foo"))
	assert.NoError(t, git.RemoveRemote(ctx, "foo"))
	assert.Error(t, git.RemoveRemote(ctx, "foo"))

	gitdir2 := filepath.Join(td, "git2")
	require.NoError(t, os.Mkdir(gitdir2, 0755))

	git, err = Clone(ctx, gitdir, gitdir2)
	require.NoError(t, err)
	require.NotNil(t, git)
	assert.Equal(t, "git", git.Name())

	tf = filepath.Join(gitdir2, "some-other-file")
	require.NoError(t, ioutil.WriteFile(tf, []byte("foobar"), 0644))
	assert.NoError(t, git.Add(ctx, "some-other-file"))
	assert.NoError(t, git.Commit(ctx, "added some-other-file"))

	revs, err := git.Revisions(ctx, "some-other-file")
	require.NoError(t, err)
	assert.Equal(t, true, len(revs) == 1)

	content, err := git.GetRevision(ctx, "some-other-file", revs[0].Hash)
	require.NoError(t, err)
	assert.Equal(t, "foobar", string(content))
}
