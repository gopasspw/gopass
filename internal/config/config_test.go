package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) { //nolint:paralleltest
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

	assert.Equal(t, []string{"core.autosync", "core.bool", "core.cliptimeout", "core.exportkeys", "core.int", "core.notifications", "core.string", "env.string", "mounts.path"}, cfg.Keys(""))

	ctx := cfg.WithConfig(context.Background())
	assert.Equal(t, true, Bool(ctx, "core.bool"))
	assert.Equal(t, "foo", String(ctx, "core.string"))
	assert.Equal(t, 42, Int(ctx, "core.int"))
}
