//go:build !windows && !darwin

package gitfs

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitSSHCommandFor(t *testing.T) {
	t.Parallel()

	writableDir := t.TempDir()
	missingDir := filepath.Join(writableDir, "does-not-exist")
	regularFile := filepath.Join(writableDir, "not-a-dir")
	require.NoError(t, os.WriteFile(regularFile, []byte("x"), 0o600))

	const wantCmd = "ssh -oControlMaster=auto -oControlPersist=600 -oControlPath=${TMPDIR:-/tmp}/.ssh-%C"

	t.Run("linux writable tmpdir", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, wantCmd, gitSSHCommandFor("linux", "", writableDir))
	})

	t.Run("android disabled", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, gitSSHCommandFor("android", "", writableDir))
	})

	t.Run("termux disabled", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, gitSSHCommandFor("linux", "0.118.0", writableDir))
	})

	t.Run("non-existent tmpdir disabled", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, gitSSHCommandFor("linux", "", missingDir))
	})

	t.Run("file as tmpdir disabled", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, gitSSHCommandFor("linux", "", regularFile))
	})

	t.Run("read-only tmpdir disabled", func(t *testing.T) {
		t.Parallel()

		if os.Geteuid() == 0 {
			t.Skip("root can write to read-only directories")
		}

		roDir := filepath.Join(t.TempDir(), "ro")
		require.NoError(t, os.Mkdir(roDir, 0o500))
		t.Cleanup(func() {
			_ = os.Chmod(roDir, 0o700)
		})

		assert.Empty(t, gitSSHCommandFor("linux", "", roDir))
	})
}

func TestGitConfigSSHCommand(t *testing.T) {
	const wantCmd = "ssh -oControlMaster=auto -oControlPersist=600 -oControlPath=${TMPDIR:-/tmp}/.ssh-%C"

	setupEnv := func(t *testing.T, termuxVersion, tmpDir string) {
		t.Helper()

		home := t.TempDir()
		t.Setenv("HOME", home)
		t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
		t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(home, "empty-gitconfig"))
		t.Setenv("GIT_CONFIG_SYSTEM", filepath.Join(home, "empty-sysconfig"))
		t.Setenv("TERMUX_VERSION", termuxVersion)
		t.Setenv("TMPDIR", tmpDir)
	}

	initRepo := func(t *testing.T, baseDir string) *Git {
		t.Helper()

		gitdir := filepath.Join(baseDir, "git")
		require.NoError(t, os.Mkdir(gitdir, 0o755))

		ctx := config.NewContextInMemory()
		ctx = ctxutil.WithAlwaysYes(ctx, true)

		buf := &bytes.Buffer{}
		prevOut := out.Stdout
		out.Stdout = buf
		t.Cleanup(func() {
			out.Stdout = prevOut
		})

		git, err := Init(ctx, gitdir, "Dead Beef", "dead.beef@example.org")
		require.NoError(t, err)

		return git
	}

	t.Run("sets dynamic TMPDIR sshCommand and preserves user override", func(t *testing.T) {
		baseDir := t.TempDir()
		setupEnv(t, "", t.TempDir())
		git := initRepo(t, baseDir)
		ctx := config.NewContextInMemory()

		got, err := git.ConfigGet(ctx, "core.sshCommand")
		require.NoError(t, err)
		assert.Equal(t, wantCmd, got)

		const customCmd = "ssh -i /custom/key"
		require.NoError(t, git.ConfigSet(ctx, "core.sshCommand", customCmd))
		require.NoError(t, git.InitConfig(ctx, "Dead Beef", "dead.beef@example.org"))

		got, err = git.ConfigGet(ctx, "core.sshCommand")
		require.NoError(t, err)
		assert.Equal(t, customCmd, got)
	})

	t.Run("skips sshCommand on Termux", func(t *testing.T) {
		baseDir := t.TempDir()
		setupEnv(t, "0.118.0", t.TempDir())
		git := initRepo(t, baseDir)
		ctx := config.NewContextInMemory()

		got, err := git.ConfigGet(ctx, "core.sshCommand")
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("skips sshCommand when TMPDIR does not exist", func(t *testing.T) {
		baseDir := t.TempDir()
		missing := filepath.Join(t.TempDir(), "missing")
		setupEnv(t, "", missing)
		git := initRepo(t, baseDir)
		ctx := config.NewContextInMemory()

		got, err := git.ConfigGet(ctx, "core.sshCommand")
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}
