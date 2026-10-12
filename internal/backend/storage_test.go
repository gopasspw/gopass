package backend

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectStorage(t *testing.T) {
	ctx := config.NewContextInMemory()

	td := t.TempDir()

	// all tests involving age should set GOPASS_HOMEDIR
	t.Setenv("GOPASS_HOMEDIR", td)
	ctx = ctxutil.WithAgePassphrase(ctx, "gopass")

	fsDir := filepath.Join(td, "fs")
	require.NoError(t, os.MkdirAll(fsDir, 0o700))

	t.Run("detect fs", func(t *testing.T) {
		r, err := DetectStorage(ctx, fsDir)
		require.NoError(t, err)
		assert.NotNil(t, r)
		assert.Equal(t, "fs", r.Name())
	})
}

func TestGitBackendFromEnv(t *testing.T) {
	for _, tc := range []struct {
		env  string
		want StorageBackend
		ok   bool
	}{
		{"", FS, false},
		{"gogit", GoGit, true},
		{"GOGIT", GoGit, true},
		{"gitfs", GitFS, true},
		{"cli", GitFS, true},
		{"cmd", GitFS, true},
		{"bogus", FS, false},
	} {
		t.Run(tc.env, func(t *testing.T) {
			t.Setenv("GOPASS_GIT_BACKEND", tc.env)
			got, ok := GitBackendFromEnv()
			assert.Equal(t, tc.ok, ok)
			if tc.ok {
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

func TestResolveStorageBackend(t *testing.T) {
	ctx := config.NewContextInMemory()

	// non-git backends are returned unchanged.
	assert.Equal(t, FS, ResolveStorageBackend(ctx, FS))
	assert.Equal(t, GoGit, ResolveStorageBackend(ctx, GoGit))

	// GitFS resolves to GitFS when git is available, otherwise to GoGit.
	got := ResolveStorageBackend(ctx, GitFS)
	if _, err := exec.LookPath("git"); err == nil {
		assert.Equal(t, GitFS, got)
	} else {
		assert.Equal(t, GoGit, got)
	}
}

func TestResolveGitBackend(t *testing.T) {
	ctx := config.NewContextInMemory()

	// an explicit context backend wins over the environment.
	t.Setenv("GOPASS_GIT_BACKEND", "gogit")
	ctx = WithStorageBackend(ctx, GitFS)
	assert.Equal(t, GitFS, ResolveGitBackend(ctx, GitFS))

	// the environment wins over the default.
	ctx = config.NewContextInMemory()
	assert.Equal(t, GoGit, ResolveGitBackend(ctx, GitFS))

	// the default is used when nothing is set.
	t.Setenv("GOPASS_GIT_BACKEND", "")
	assert.Equal(t, ResolveStorageBackend(ctx, GitFS), ResolveGitBackend(ctx, GitFS))
}

func TestGitBackendFromEnvOrConfig(t *testing.T) {
	ctx := config.NewContextInMemory()

	// environment wins.
	t.Setenv("GOPASS_GIT_BACKEND", "gogit")
	assert.Equal(t, GoGit, GitBackendFromEnvOrConfig(ctx, GitFS))

	// config is used when the environment is unset.
	t.Setenv("GOPASS_GIT_BACKEND", "")
	cfg := config.NewInMemory()
	require.NoError(t, cfg.Set("", "storage.backend", "gogit"))
	ctx = cfg.WithConfig(context.Background())
	assert.Equal(t, GoGit, GitBackendFromEnvOrConfig(ctx, GitFS))

	// a non-git-family config value is ignored: clone/rcs init must not resolve
	// to FS (whose Clone only creates an empty directory).
	cfg = config.NewInMemory()
	require.NoError(t, cfg.Set("", "storage.backend", "fs"))
	ctx = cfg.WithConfig(context.Background())
	assert.Equal(t, ResolveStorageBackend(ctx, GitFS), GitBackendFromEnvOrConfig(ctx, GitFS))

	// default is used when neither is set.
	ctx = config.NewContextInMemory()
	assert.Equal(t, ResolveStorageBackend(ctx, GitFS), GitBackendFromEnvOrConfig(ctx, GitFS))
}

func TestResolveGitBackendIgnoresNonGitConfig(t *testing.T) {
	cfg := config.NewInMemory()
	require.NoError(t, cfg.Set("", "storage.backend", "fs"))
	ctx := cfg.WithConfig(context.Background())

	// a non-git-family config value must not override the git-family default.
	assert.Equal(t, ResolveStorageBackend(ctx, GitFS), ResolveGitBackend(ctx, GitFS))
}
