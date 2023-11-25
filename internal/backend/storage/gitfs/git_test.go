package gitfs

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGit(t *testing.T) {
	td := t.TempDir()

	gitdir := filepath.Join(td, "git")
	require.NoError(t, os.Mkdir(gitdir, 0o755))
	gitdir2 := filepath.Join(td, "git2")
	require.NoError(t, os.Mkdir(gitdir2, 0o755))

	ctx := config.NewNoWrites().WithConfig(context.Background())
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
		require.NoError(t, os.WriteFile(tf, []byte("foobar"), 0o644))
		require.NoError(t, git.Add(ctx, "some-file"))
		assert.True(t, git.HasStagedChanges(ctx))
		require.NoError(t, git.Commit(ctx, "added some-file"))
		assert.False(t, git.HasStagedChanges(ctx))

		require.Error(t, git.Push(ctx, "origin", "master"))
		require.Error(t, git.Pull(ctx, "origin", "master"))
	})

	t.Run("open existing repo", func(t *testing.T) {
		git, err := New(gitdir)
		require.NoError(t, err)
		require.NotNil(t, git)
		assert.Equal(t, "gitfs", git.Name())
		require.NoError(t, git.AddRemote(ctx, "foo", "file:///tmp/foo"))
		require.NoError(t, git.RemoveRemote(ctx, "foo"))
		require.Error(t, git.RemoveRemote(ctx, "foo"))
	})

	t.Run("clone existing repo", func(t *testing.T) {
		git, err := Clone(ctx, gitdir, gitdir2, "", "")
		require.NoError(t, err)
		require.NotNil(t, git)
		assert.Equal(t, "gitfs", git.Name())

		tf := filepath.Join(gitdir2, "some-other-file")
		require.NoError(t, os.WriteFile(tf, []byte("foobar"), 0o644))
		require.NoError(t, git.Add(ctx, "some-other-file"))
		require.NoError(t, git.Commit(ctx, "added some-other-file"))

		revs, err := git.Revisions(ctx, "some-other-file")
		require.NoError(t, err)
		assert.Len(t, revs, 1)

		content, err := git.GetRevision(ctx, "some-other-file", revs[0].Hash)
		require.NoError(t, err)
		assert.Equal(t, "foobar", string(content))
	})
}

func TestParseVersion(t *testing.T) {
	for _, tc := range []struct {
		name    string
		in      string
		sv      semver.Version
		wantErr bool
	}{
		{
			name:    "empty",
			in:      "",
			wantErr: true,
		},
		{
			name:    "invalid",
			in:      "foo",
			wantErr: true,
		},
		{
			name: "valid",
			in:   "2.30.0",
			sv:   semver.MustParse("2.30.0"),
		},
		{
			name: "invalid-recovered", // GH-2686
			in:   "2.42.0.windows.2",
			sv:   semver.MustParse("2.42.0"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sv, err := parseVersion(tc.in)
			assert.Equal(t, tc.sv, sv)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
