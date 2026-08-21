package age

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	ctx := t.Context()
	a, err := New(ctx, false, "")
	require.NoError(t, err)
	assert.NotNil(t, a)
}

func TestAgentSpawnGuard(t *testing.T) {
	// By default this process is not an agent-spawn child.
	require.False(t, isAgentSpawnProcess())
	// The guard env short-circuits tryStartAgent when set.
	t.Setenv(SpawnGuardEnv, "1")
	require.True(t, isAgentSpawnProcess())
	// Contract: the env var name is stable (the CLI launcher sets exactly this).
	assert.Equal(t, "GOPASS_AGE_AGENT_SPAWNING", SpawnGuardEnv)
}

// TestTryStartAgentGuardSkips verifies that the spawn guard short-circuits
// tryStartAgent before the launcher is consulted, even when a launcher is
// registered and the agent is enabled. Exercises tryStartAgent through New
// (which calls it internally).
func TestTryStartAgentGuardSkips(t *testing.T) {
	called := false
	launcher := func(context.Context) error {
		called = true

		return nil
	}
	t.Setenv(SpawnGuardEnv, "1")
	ctx := WithAgentLauncher(ctxWithAgentEnabled(t), launcher)

	a, err := New(ctx, "")
	require.NoError(t, err)
	_ = a
	require.False(t, called, "guard must short-circuit before the launcher is called")
}

// TestTryStartAgentNoLauncherSkips verifies the library-consumer path: with the
// agent enabled but no launcher registered, tryStartAgent must skip without
// spawning and without hanging on the 3s ping backoff. After this change the
// backend has no spawn code, so this also guards against any future re-introduced
// default spawn.
//
// GOPASS_HOMEDIR is isolated to a temp dir so the agent socket resolves
// somewhere no agent is listening; otherwise, on a developer machine with a
// live gopass-age-agent, Ping() would succeed and this test would pass via the
// "already running" branch instead of the no-launcher branch it claims to test.
func TestTryStartAgentNoLauncherSkips(t *testing.T) {
	t.Setenv("GOPASS_HOMEDIR", t.TempDir())
	ctx := ctxWithAgentEnabled(t) // no launcher registered

	done := make(chan error, 1)
	go func() {
		_, err := New(ctx, "")
		done <- err
	}()
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("New hung; expected no-launcher path to skip before the 3s backoff")
	}
}

// TestTryStartAgentLauncherError verifies that a launcher returning an error is
// handled gracefully: tryStartAgent logs and returns without falling through to
// the 3s ping backoff.
func TestTryStartAgentLauncherError(t *testing.T) {
	t.Setenv("GOPASS_HOMEDIR", t.TempDir()) // isolate socket so Ping fails, forcing the launcher path
	launcher := func(context.Context) error { return errors.New("kaboom") }
	ctx := WithAgentLauncher(ctxWithAgentEnabled(t), launcher)

	done := make(chan error, 1)
	go func() {
		_, err := New(ctx, "")
		done <- err
	}()
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("New hung; expected launcher-error path to return before the 3s backoff")
	}
}

func TestNewExpandsSshKeyPathTilde(t *testing.T) {
	td := t.TempDir()
	t.Setenv("GOPASS_HOMEDIR", td)

	// A leading ~/ is expanded to the home dir (appdir.UserHome honors GOPASS_HOMEDIR).
	a, err := New(t.Context(), false, "~/custom-ssh")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(td, "custom-ssh"), a.sshKeyPath)

	// Absolute paths are left untouched by fsutil.ExpandHomedir.
	a2, err := New(t.Context(), false, "/opt/keys")
	require.NoError(t, err)
	assert.Equal(t, "/opt/keys", a2.sshKeyPath)

	// ~user/ and a bare ~ are NOT expanded: ExpandHomedir only matches the ~/ prefix.
	a3, err := New(t.Context(), false, "~alice/keys")
	require.NoError(t, err)
	assert.Equal(t, "~alice/keys", a3.sshKeyPath)

	// Empty stays empty (the len>1 guard inside ExpandHomedir).
	a4, err := New(t.Context(), false, "")
	require.NoError(t, err)
	assert.Empty(t, a4.sshKeyPath)
}

func TestInitialized(t *testing.T) {
	ctx := t.Context()
	a, err := New(ctx, false, "")
	require.NoError(t, err)
	assert.NotNil(t, a)

	err = a.Initialized(ctx)
	require.NoError(t, err)
}

func TestName(t *testing.T) {
	ctx := t.Context()
	a, err := New(ctx, false, "")
	require.NoError(t, err)
	assert.NotNil(t, a)

	name := a.Name()
	assert.Equal(t, "age", name)
}

func TestVersion(t *testing.T) {
	ctx := t.Context()
	a, err := New(ctx, false, "")
	require.NoError(t, err)
	assert.NotNil(t, a)

	version := a.Version(ctx)
	expectedVersion := debug.ModuleVersion("filippo.io/age")
	assert.Equal(t, expectedVersion, version)
}

func TestExt(t *testing.T) {
	ctx := t.Context()
	a, err := New(ctx, false, "")
	require.NoError(t, err)
	assert.NotNil(t, a)

	ext := a.Ext()
	assert.Equal(t, Ext, ext)
}

func TestIDFile(t *testing.T) {
	ctx := t.Context()
	a, err := New(ctx, false, "")
	require.NoError(t, err)
	assert.NotNil(t, a)

	idFile := a.IDFile()
	assert.Equal(t, IDFile, idFile)
}

func TestConcurrency(t *testing.T) {
	ctx := t.Context()
	a, err := New(ctx, false, "")
	require.NoError(t, err)
	assert.NotNil(t, a)

	concurrency := a.Concurrency()
	assert.Equal(t, 1, concurrency)
}
