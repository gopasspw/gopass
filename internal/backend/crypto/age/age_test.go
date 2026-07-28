package age

import (
	"path/filepath"
	"testing"

	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	ctx := t.Context()
	a, err := New(ctx, "")
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

func TestNewExpandsSshKeyPathTilde(t *testing.T) {
	td := t.TempDir()
	t.Setenv("GOPASS_HOMEDIR", td)

	// A leading ~/ is expanded to the home dir (appdir.UserHome honors GOPASS_HOMEDIR).
	a, err := New(t.Context(), "~/custom-ssh")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(td, "custom-ssh"), a.sshKeyPath)

	// Absolute paths are left untouched by fsutil.ExpandHomedir.
	a2, err := New(t.Context(), "/opt/keys")
	require.NoError(t, err)
	assert.Equal(t, "/opt/keys", a2.sshKeyPath)

	// ~user/ and a bare ~ are NOT expanded: ExpandHomedir only matches the ~/ prefix.
	a3, err := New(t.Context(), "~alice/keys")
	require.NoError(t, err)
	assert.Equal(t, "~alice/keys", a3.sshKeyPath)

	// Empty stays empty (the len>1 guard inside ExpandHomedir).
	a4, err := New(t.Context(), "")
	require.NoError(t, err)
	assert.Empty(t, a4.sshKeyPath)
}

func TestInitialized(t *testing.T) {
	ctx := t.Context()
	a, err := New(ctx, "")
	require.NoError(t, err)
	assert.NotNil(t, a)

	err = a.Initialized(ctx)
	require.NoError(t, err)
}

func TestName(t *testing.T) {
	ctx := t.Context()
	a, err := New(ctx, "")
	require.NoError(t, err)
	assert.NotNil(t, a)

	name := a.Name()
	assert.Equal(t, "age", name)
}

func TestVersion(t *testing.T) {
	ctx := t.Context()
	a, err := New(ctx, "")
	require.NoError(t, err)
	assert.NotNil(t, a)

	version := a.Version(ctx)
	expectedVersion := debug.ModuleVersion("filippo.io/age")
	assert.Equal(t, expectedVersion, version)
}

func TestExt(t *testing.T) {
	ctx := t.Context()
	a, err := New(ctx, "")
	require.NoError(t, err)
	assert.NotNil(t, a)

	ext := a.Ext()
	assert.Equal(t, Ext, ext)
}

func TestIDFile(t *testing.T) {
	ctx := t.Context()
	a, err := New(ctx, "")
	require.NoError(t, err)
	assert.NotNil(t, a)

	idFile := a.IDFile()
	assert.Equal(t, IDFile, idFile)
}

func TestConcurrency(t *testing.T) {
	ctx := t.Context()
	a, err := New(ctx, "")
	require.NoError(t, err)
	assert.NotNil(t, a)

	concurrency := a.Concurrency()
	assert.Equal(t, 1, concurrency)
}
