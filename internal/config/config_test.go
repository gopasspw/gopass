package config

import (
	"context"
	"testing"

	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	u := gptest.NewUnitTester(t)
	assert.NotNil(t, u)

	td := t.TempDir()
	t.Setenv("GOPASS_HOMEDIR", td)

	// this will write to the tempdir
	cfg := New()

	assert.False(t, cfg.IsSet("core.string"))
	assert.NoError(t, cfg.Set("", "core.string", "foo"))
	assert.NoError(t, cfg.Set("", "core.bool", "true"))
	assert.NoError(t, cfg.Set("", "core.int", "42"))

	assert.Equal(t, "foo", cfg.Get("core.string"))
	assert.Equal(t, true, cfg.GetBool("core.bool"))
	assert.Equal(t, 42, cfg.GetInt("core.int"))

	assert.NoError(t, cfg.SetEnv("env.string", "foo"))
	assert.Equal(t, "foo", cfg.Get("env.string"))

	assert.Equal(t, []string{"core.autopush", "core.autosync", "core.bool", "core.cliptimeout", "core.exportkeys", "core.int", "core.notifications", "core.string", "env.string", "mounts.path"}, cfg.Keys(""))

	ctx := cfg.WithConfig(context.Background())
	assert.Equal(t, true, Bool(ctx, "core.bool"))
	assert.Equal(t, "foo", String(ctx, "core.string"))
	assert.Equal(t, 42, Int(ctx, "core.int"))
}

func TestEnvConfig(t *testing.T) {
	envs := map[string]string{
		"GOPASS_CONFIG_CONFIG_COUNT":   "1",
		"GOPASS_CONFIG_CONFIG_KEY_0":   "core.autosync",
		"GOPASS_CONFIG_CONFIG_VALUE_0": "false",
	}
	for k, v := range envs {
		t.Setenv(k, v)
	}

	u := gptest.NewUnitTester(t)
	assert.NotNil(t, u)

	td := t.TempDir()
	t.Setenv("GOPASS_HOMEDIR", td)

	// this will write to the tempdir
	cfg := New()

	assert.Equal(t, "false", cfg.Get("core.autosync"))
}

func TestInvalidEnvConfig(t *testing.T) {
	envs := map[string]string{
		"GOPASS_CONFIG__CONFIG_COUNT":   "1",
		"GOPASS_CONFIG__CONFIG_KEY_0":   "core.autosync",
		"GOPASS_CONFIG__CONFIG_VALUE_0": "false",
	}
	for k, v := range envs {
		t.Setenv(k, v)
	}

	u := gptest.NewUnitTester(t)
	assert.NotNil(t, u)

	td := t.TempDir()
	t.Setenv("GOPASS_HOMEDIR", td)

	// this will write to the tempdir
	cfg := New()

	assert.Equal(t, "true", cfg.Get("core.autosync"))
}
