package gitconfig

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigs(t *testing.T) {
	td := t.TempDir()

	t.Setenv("GOPASS_HOMEDIR", td)

	c := New()
	c.SystemConfig = filepath.Join(td, "system")
	c.GlobalConfig = "global"
	c.LocalConfig = "local"
	c.WorktreeConfig = "worktree"
	c.EnvPrefix = "GPTEST_CONFIG"

	require.NoError(t, os.WriteFile(c.SystemConfig, []byte(`[system]
	key = system
`), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(td, c.GlobalConfig), []byte(`[global]
	key = global
[alias "foo"]
	key = bar
`), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(td, c.LocalConfig), []byte(`[local]
	key = local
`), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(td, c.WorktreeConfig), []byte(`[worktree]
	key = worktree
`), 0o600))
	t.Setenv("GPTEST_CONFIG_COUNT", "1")
	t.Setenv("GPTEST_CONFIG_KEY_0", "env.key")
	t.Setenv("GPTEST_CONFIG_VALUE_0", "env")

	// Load the configs
	c.LoadAll(td)

	assert.True(t, c.HasGlobalConfig())

	// Read the expected keys
	assert.Equal(t, "system", c.Get("system.key"))
	assert.Equal(t, "global", c.Get("global.key"))
	assert.Equal(t, "local", c.Get("local.key"))
	assert.Equal(t, "worktree", c.Get("worktree.key"))
	assert.Equal(t, "env", c.Get("env.key"))

	assert.Equal(t, "global", c.GetGlobal("global.key"))
	assert.Equal(t, "", c.GetGlobal("local.key"))

	assert.Equal(t, "local", c.GetLocal("local.key"))
	assert.Equal(t, "", c.GetLocal("global.key"))

	for _, k := range []string{"system.key", "global.key", "local.key", "worktree.key", "env.key"} {
		assert.True(t, c.IsSet(k))
	}

	// SetLocal
	require.NoError(t, c.SetLocal("global.fakekey", "local"))
	assert.Equal(t, "local", c.GetLocal("global.fakekey"))
	assert.Equal(t, "", c.GetGlobal("global.fakekey"))
	require.NoError(t, c.UnsetLocal("global.fakekey"))
	assert.Equal(t, "", c.Get("global.fakekey"))

	// SetGlobal
	require.NoError(t, c.SetGlobal("local.fakekey", "global"))
	assert.Equal(t, "", c.GetLocal("local.fakekey"))
	assert.Equal(t, "global", c.GetGlobal("local.fakekey"))
	require.NoError(t, c.UnsetGlobal("local.fakekey"))
	assert.Equal(t, "", c.Get("local.fakekey"))

	// SetEnv
	require.NoError(t, c.SetEnv("worktree.fakekey", "env"))
	assert.Equal(t, "", c.GetLocal("worktree.fakekey"))
	assert.Equal(t, "", c.GetGlobal("worktree.fakekey"))
	assert.Equal(t, "env", c.Get("worktree.fakekey"))

	// List
	assert.Equal(t, []string{"alias.foo.key", "env.key", "global.key", "local.key", "system.key", "worktree.fakekey", "worktree.key"}, c.Keys())
	assert.Equal(t, []string{"global.key"}, c.List("global."))
	assert.Equal(t, []string{"alias", "env", "global", "local", "system", "worktree"}, c.ListSections())
	assert.Equal(t, []string{"foo"}, c.ListSubsections("alias"))

	// Failure modes
	c.workdir = ""
	c.local = nil
	c.global = nil
	c.env = nil

	require.Error(t, c.SetLocal("core.foo", "bar"))
	require.NoError(t, c.SetGlobal("core.global", "foo"))
	require.NoError(t, c.SetEnv("env.var", "var"))
}
