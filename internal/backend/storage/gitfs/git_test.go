package gitfs

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGit(t *testing.T) {
	td, err := os.MkdirTemp("", "gopass-")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()

	gitdir := filepath.Join(td, "git")
	require.NoError(t, os.Mkdir(gitdir, 0755))
	gitdir2 := filepath.Join(td, "git2")
	require.NoError(t, os.Mkdir(gitdir2, 0755))

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	t.Run("init new repo", func(t *testing.T) {
		git, err := Init(ctx, gitdir, "Dead Beef", "dead.beef@example.org")
		require.NoError(t, err)
		require.NotNil(t, git)

		sv := git.Version(ctx)
		assert.NotEqual(t, "", sv.String())

		assert.True(t, git.IsInitialized())
		tf := filepath.Join(gitdir, "some-file")
		require.NoError(t, os.WriteFile(tf, []byte("foobar"), 0644))
		assert.NoError(t, git.Add(ctx, "some-file"))
		assert.True(t, git.HasStagedChanges(ctx))
		assert.NoError(t, git.Commit(ctx, "added some-file"))
		assert.False(t, git.HasStagedChanges(ctx))

		assert.Error(t, git.Push(ctx, "origin", "master"))
		assert.Error(t, git.Pull(ctx, "origin", "master"))
	})

	t.Run("open existing repo", func(t *testing.T) {
		git, err := New(gitdir)
		require.NoError(t, err)
		require.NotNil(t, git)
		assert.Equal(t, "git", git.Name())
		assert.NoError(t, git.AddRemote(ctx, "foo", "file:///tmp/foo"))
		assert.NoError(t, git.RemoveRemote(ctx, "foo"))
		assert.Error(t, git.RemoveRemote(ctx, "foo"))
	})

	t.Run("clone existing repo", func(t *testing.T) {
		git, err := Clone(ctx, gitdir, gitdir2, "", "")
		require.NoError(t, err)
		require.NotNil(t, git)
		assert.Equal(t, "git", git.Name())

		tf := filepath.Join(gitdir2, "some-other-file")
		require.NoError(t, os.WriteFile(tf, []byte("foobar"), 0644))
		assert.NoError(t, git.Add(ctx, "some-other-file"))
		assert.NoError(t, git.Commit(ctx, "added some-other-file"))

		revs, err := git.Revisions(ctx, "some-other-file")
		require.NoError(t, err)
		assert.True(t, len(revs) == 1)

		content, err := git.GetRevision(ctx, "some-other-file", revs[0].Hash)
		require.NoError(t, err)
		assert.Equal(t, "foobar", string(content))
	})
}
