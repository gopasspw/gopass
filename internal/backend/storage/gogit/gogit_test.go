package gogit

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testContext() context.Context {
	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithGitInit(ctx, true)

	return ctx
}

func TestGoGit(t *testing.T) {
	td := t.TempDir()

	gitdir := filepath.Join(td, "git")
	require.NoError(t, os.Mkdir(gitdir, 0o755))
	gitdir2 := filepath.Join(td, "git2")
	require.NoError(t, os.Mkdir(gitdir2, 0o755))

	ctx := testContext()

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	t.Run("init new repo", func(t *testing.T) {
		g, err := Init(ctx, gitdir, "Dead Beef", "dead.beef@example.org")
		require.NoError(t, err)
		require.NotNil(t, g)

		assert.True(t, g.IsInitialized())
		assert.Equal(t, "gogit", g.Name())
		assert.NotEmpty(t, g.Version(ctx).String())

		tf := filepath.Join(gitdir, "some-file")
		require.NoError(t, os.WriteFile(tf, []byte("foobar"), 0o644))
		require.NoError(t, g.Add(ctx, "some-file"))
		assert.True(t, g.HasStagedChanges(ctx))
		require.NoError(t, g.Commit(ctx, "added some-file"))
		assert.False(t, g.HasStagedChanges(ctx))

		require.ErrorIs(t, g.Push(ctx, "origin", "master"), store.ErrGitNoRemote)
		require.ErrorIs(t, g.Pull(ctx, "origin", "master"), store.ErrGitNoRemote)
	})

	t.Run("open existing repo", func(t *testing.T) {
		g, err := New(gitdir)
		require.NoError(t, err)
		require.NotNil(t, g)
		assert.Equal(t, "gogit", g.Name())
		require.NoError(t, g.AddRemote(ctx, "foo", "file:///tmp/foo"))
		require.NoError(t, g.RemoveRemote(ctx, "foo"))
		require.Error(t, g.RemoveRemote(ctx, "foo"))
	})

	t.Run("clone existing repo", func(t *testing.T) {
		g, err := Clone(ctx, gitdir, gitdir2, "", "")
		require.NoError(t, err)
		require.NotNil(t, g)
		assert.Equal(t, "gogit", g.Name())

		tf := filepath.Join(gitdir2, "some-other-file")
		require.NoError(t, os.WriteFile(tf, []byte("foobar"), 0o644))
		require.NoError(t, g.Add(ctx, "some-other-file"))
		require.NoError(t, g.Commit(ctx, "added some-other-file"))

		revs, err := g.Revisions(ctx, "some-other-file")
		require.NoError(t, err)
		require.Len(t, revs, 1)
		assert.Equal(t, "added some-other-file", revs[0].Subject)

		content, err := g.GetRevision(ctx, "some-other-file", revs[0].Hash)
		require.NoError(t, err)
		assert.Equal(t, "foobar", string(content))

		content, err = g.GetRevision(ctx, "some-other-file", "latest")
		require.NoError(t, err)
		assert.Equal(t, "foobar", string(content))
	})

	t.Run("file operations", func(t *testing.T) {
		g, err := New(gitdir)
		require.NoError(t, err)

		require.NoError(t, g.Set(ctx, "dir/secret", []byte("s3cr3t")))
		assert.True(t, g.Exists(ctx, "dir/secret"))
		got, err := g.Get(ctx, "dir/secret")
		require.NoError(t, err)
		assert.Equal(t, "s3cr3t", string(got))
		assert.True(t, g.IsDir(ctx, "dir"))

		require.NoError(t, g.Move(ctx, "dir/secret", "dir/moved", true))
		assert.False(t, g.Exists(ctx, "dir/secret"))
		assert.True(t, g.Exists(ctx, "dir/moved"))

		require.NoError(t, g.Delete(ctx, "dir/moved"))
		assert.False(t, g.Exists(ctx, "dir/moved"))
	})

	t.Run("status", func(t *testing.T) {
		g, err := New(gitdir)
		require.NoError(t, err)
		require.NoError(t, g.Set(ctx, "untracked-secret", []byte("x")))
		st, err := g.Status(ctx)
		require.NoError(t, err)
		assert.Contains(t, string(st), "untracked-secret")
	})

	t.Run("compact", func(t *testing.T) {
		g, err := New(gitdir)
		require.NoError(t, err)
		require.NoError(t, g.Compact(ctx))
	})
}

func TestGoGitErrors(t *testing.T) {
	ctx := testContext()
	td := t.TempDir()

	g, err := Init(ctxutil.WithGitInit(ctx, false), td, "", "")
	require.NoError(t, err)

	// no commits yet -> nothing to commit
	require.ErrorIs(t, g.Commit(ctx, "empty"), store.ErrGitNothingToCommit)
	require.NoError(t, g.TryCommit(ctx, "empty"))

	// no remote
	require.ErrorIs(t, g.Push(ctx, "", ""), store.ErrGitNoRemote)
	require.NoError(t, g.TryPush(ctx, "", ""))
}

func TestGoGitMergeMatrix(t *testing.T) {
	ctx := testContext()

	// setup: a "remote" bare-ish repo and a local clone.
	remoteDir := t.TempDir()
	remote, err := Init(ctx, remoteDir, "Remote", "remote@example.org")
	require.NoError(t, err)
	require.NoError(t, remote.Set(ctx, "base", []byte("base")))
	require.NoError(t, remote.Add(ctx, "base"))
	require.NoError(t, remote.Commit(ctx, "base"))

	localDir := t.TempDir()
	local, err := Clone(ctx, remoteDir, localDir, "Local", "local@example.org")
	require.NoError(t, err)

	// local change only
	require.NoError(t, local.Set(ctx, "local-only", []byte("local")))
	require.NoError(t, local.Add(ctx, "local-only"))
	require.NoError(t, local.Commit(ctx, "local-only"))

	// remote change only
	require.NoError(t, remote.Set(ctx, "remote-only", []byte("remote")))
	require.NoError(t, remote.Add(ctx, "remote-only"))
	require.NoError(t, remote.Commit(ctx, "remote-only"))

	// true conflict on "base"
	require.NoError(t, local.Set(ctx, "base", []byte("local-base")))
	require.NoError(t, local.Add(ctx, "base"))
	require.NoError(t, local.Commit(ctx, "local-base"))
	require.NoError(t, remote.Set(ctx, "base", []byte("remote-base")))
	require.NoError(t, remote.Add(ctx, "base"))
	require.NoError(t, remote.Commit(ctx, "remote-base"))

	require.NoError(t, local.Pull(ctx, "origin", "master"))

	// local-only change is kept
	got, err := local.Get(ctx, "local-only")
	require.NoError(t, err)
	assert.Equal(t, "local", string(got))

	// remote-only change is applied
	got, err = local.Get(ctx, "remote-only")
	require.NoError(t, err)
	assert.Equal(t, "remote", string(got))

	// conflict: local kept, remote materialised as *.conflict-<sha>
	got, err = local.Get(ctx, "base")
	require.NoError(t, err)
	assert.Equal(t, "local-base", string(got))

	conflicts, err := filepath.Glob(filepath.Join(localDir, "base.conflict-*"))
	require.NoError(t, err)
	require.Len(t, conflicts, 1)
	cb, err := os.ReadFile(conflicts[0])
	require.NoError(t, err)
	assert.Equal(t, "remote-base", string(cb))

	// merge commit has two parents
	head, err := local.repo.Head()
	require.NoError(t, err)
	commit, err := local.repo.CommitObject(head.Hash())
	require.NoError(t, err)
	assert.Equal(t, 2, commit.NumParents())
}

func TestGoGitDeleteVsModify(t *testing.T) {
	ctx := testContext()

	remoteDir := t.TempDir()
	remote, err := Init(ctx, remoteDir, "Remote", "remote@example.org")
	require.NoError(t, err)
	require.NoError(t, remote.Set(ctx, "keep", []byte("original")))
	require.NoError(t, remote.Add(ctx, "keep"))
	require.NoError(t, remote.Commit(ctx, "keep"))

	localDir := t.TempDir()
	local, err := Clone(ctx, remoteDir, localDir, "Local", "local@example.org")
	require.NoError(t, err)

	// local deletes, remote modifies -> modification must win (never drop a secret)
	require.NoError(t, local.Delete(ctx, "keep"))
	require.NoError(t, local.Add(ctx, "keep"))
	require.NoError(t, local.Commit(ctx, "delete keep"))

	require.NoError(t, remote.Set(ctx, "keep", []byte("modified")))
	require.NoError(t, remote.Add(ctx, "keep"))
	require.NoError(t, remote.Commit(ctx, "modify keep"))

	require.NoError(t, local.Pull(ctx, "origin", "master"))

	got, err := local.Get(ctx, "keep")
	require.NoError(t, err)
	assert.Equal(t, "modified", string(got))
}

func TestConflictName(t *testing.T) {
	h := plumbing.NewHash("1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b")
	assert.Equal(t, "secrets/api.conflict-1a2b3c4.gpg", conflictName("secrets/api.gpg", h))
	assert.Equal(t, "secrets/api.conflict-1a2b3c4", conflictName("secrets/api", h))
}

func TestSplitKey(t *testing.T) {
	for _, tc := range []struct {
		in            string
		sec, sub, opt string
	}{
		{"user.name", "user", "", "name"},
		{"branch.master.remote", "branch", "master", "remote"},
		{"remote.origin.url", "remote", "origin", "url"},
	} {
		sec, sub, opt := splitKey(tc.in)
		assert.Equal(t, tc.sec, sec, tc.in)
		assert.Equal(t, tc.sub, sub, tc.in)
		assert.Equal(t, tc.opt, opt, tc.in)
	}
}
