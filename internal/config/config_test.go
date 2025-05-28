package config

import (
	"testing"

	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	u := gptest.NewUnitTester(t)
	assert.NotNil(t, u)

	td := t.TempDir()
	t.Setenv("GOPASS_HOMEDIR", td)

	// this will write to the tempdir
	cfg := New()

	assert.False(t, cfg.IsSet("core.string"))
	require.NoError(t, cfg.Set("", "core.string", "foo"))
	require.NoError(t, cfg.Set("", "core.bool", "true"))
	require.NoError(t, cfg.Set("", "core.int", "42"))

	assert.Equal(t, "foo", cfg.Get("core.string"))
	assert.True(t, AsBool(cfg.Get("core.bool")))
	assert.Equal(t, 42, AsInt(cfg.Get("core.int")))

	require.NoError(t, cfg.SetEnv("env.string", "foo"))
	assert.Equal(t, "foo", cfg.Get("env.string"))

	// test default values
	assert.Equal(t, []string{"core.autopush", "core.autosync", "core.bool", "core.cliptimeout", "core.exportkeys", "core.follow-references", "core.int", "core.notifications", "core.string", "env.string", "mounts.path", "pwgen.xkcd-lang"}, cfg.Keys(""))
	for key, expected := range defaults {
		assert.Equal(t, expected, cfg.Get(key))
	}
	require.NoError(t, cfg.Set("", "pwgen.xkcd-lang", "de"))
	assert.Equal(t, "de", cfg.Get("pwgen.xkcd-lang"))

	ctx := cfg.WithConfig(t.Context())
	assert.True(t, Bool(ctx, "core.bool"))
	assert.Equal(t, "foo", String(ctx, "core.string"))
	assert.Equal(t, 42, Int(ctx, "core.int"))

	require.NoError(t, cfg.SetEnv("generate.length", "16"))
	actualLength, _ := DefaultPasswordLengthFromEnv(ctx)
	assert.Equal(t, 16, actualLength)
}

func TestEnvConfig(t *testing.T) {
	envs := map[string]string{
		"GOPASS_CONFIG_COUNT":   "2",
		"GOPASS_CONFIG_KEY_0":   "core.autosync",
		"GOPASS_CONFIG_VALUE_0": "false",
		"GOPASS_CONFIG_KEY_1":   "show.safecontent",
		"GOPASS_CONFIG_VALUE_1": "true",
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
	assert.Equal(t, "true", cfg.Get("show.safecontent"))
}

func TestInvalidEnvConfig(t *testing.T) {
	envs := map[string]string{
		// notice the double _ in the middle, this is a regression test
		"GOPASS_CONFIG__CONFIG_COUNT":   "1",
		"GOPASS_CONFIG__CONFIG_KEY_0":   "core.autosync",
		"GOPASS_CONFIG__CONFIG_VALUE_0": "false",
		// old format
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

	assert.Equal(t, "true", cfg.Get("core.autosync"))
}

func TestOptsMigration(t *testing.T) {
	t.Run("migrate global options", func(t *testing.T) {
		// we use our own temp dir
		td := t.TempDir()
		t.Setenv("GOPASS_HOMEDIR", td)

		cfg := New()
		// default config should not have been populated, we didn't call NewUnitTester
		assert.False(t, cfg.IsSet("generate.autoclip"))
		assert.False(t, cfg.IsSet("core.showsafecontent"))
		assert.False(t, cfg.IsSet("core.safecontent"))
		assert.False(t, cfg.IsSet("show.safecontent"))
		// this will write to the tempdir
		require.NoError(t, cfg.Set("", "core.showsafecontent", "true"))
		assert.True(t, cfg.IsSet("core.showsafecontent"))
		assert.Equal(t, "true", cfg.root.GetGlobal("core.showsafecontent"))
		assert.Empty(t, cfg.root.GetLocal("core.showsafecontent"))
		assert.False(t, cfg.IsSet("core.safecontent"))
		assert.False(t, cfg.IsSet("show.safecontent"))

		t.Setenv("GOPASS_CONFIG_NO_MIGRATE", "")
		// we test the migration path
		// this will read from the tempdir and should migrate the above "old option" to its new expected value
		// but only at the global level, not the local one since it was a global option
		cfg2 := New()
		assert.False(t, cfg2.IsSet("core.showsafecontent"))
		assert.False(t, cfg2.IsSet("core.safecontent"))
		assert.Empty(t, cfg2.root.GetGlobal("core.showsafecontent"))
		assert.Equal(t, "true", cfg2.root.GetGlobal("show.safecontent"))
		assert.Empty(t, cfg2.root.GetLocal("show.safecontent"))
	})

	t.Run("migrated config matches test config", func(t *testing.T) {
		u := gptest.NewUnitTester(t)
		assert.NotNil(t, u)
		cfg := New()
		assert.True(t, cfg.IsSet("core.autoimport"))
		assert.True(t, cfg.IsSet("core.cliptimeout"))
		assert.True(t, cfg.IsSet("core.notifications"))
		assert.True(t, cfg.IsSet("core.nopager"))
		assert.False(t, cfg.IsSet("core.autoclip"))
		assert.True(t, cfg.IsSet("generate.autoclip"))
		assert.False(t, cfg.IsSet("core.showsafecontent"))
		assert.False(t, cfg.IsSet("show.safecontent"))
		assert.False(t, cfg.IsSet("core.safecontent"))

		t.Setenv("GOPASS_CONFIG_NO_MIGRATE", "")
		// we test the migration path
		cfg = New()
		assert.True(t, cfg.IsSet("core.autoimport"))
		assert.True(t, cfg.IsSet("core.cliptimeout"))
		assert.True(t, cfg.IsSet("core.notifications"))
		assert.True(t, cfg.IsSet("core.nopager"))
		assert.False(t, cfg.IsSet("core.autoclip"))
		assert.True(t, cfg.IsSet("generate.autoclip"))
		assert.False(t, cfg.IsSet("core.showsafecontent"))
		assert.False(t, cfg.IsSet("show.safecontent"))
		assert.False(t, cfg.IsSet("core.safecontent"))
	})

	t.Run("migrate local options", func(t *testing.T) {
		u := gptest.NewUnitTester(t)
		assert.NotNil(t, u)

		cfg := New()
		// this will write to the local config because of the <root> arg
		require.NoError(t, cfg.Set("<root>", "core.showsafecontent", "true"))
		assert.Equal(t, "true", cfg.root.GetLocal("core.showsafecontent"))
		assert.Empty(t, cfg.root.GetGlobal("core.showsafecontent"))
		assert.False(t, cfg.IsSet("show.safecontent"))

		t.Setenv("GOPASS_CONFIG_NO_MIGRATE", "")
		// we test the migration path
		cfg = New()
		assert.False(t, cfg.IsSet("core.showsafecontent"))
		assert.Equal(t, "true", cfg.root.GetLocal("show.safecontent"))
		assert.Empty(t, cfg.root.GetGlobal("show.safecontent"))
	})

	t.Run("env variable are not migrated", func(t *testing.T) {
		envs := map[string]string{
			"GOPASS_CONFIG_COUNT":      "1",
			"GOPASS_CONFIG_KEY_0":      "core.showsafecontent",
			"GOPASS_CONFIG_VALUE_0":    "true",
			"GOPASS_CONFIG_NO_MIGRATE": "",
		}
		for k, v := range envs {
			t.Setenv(k, v)
		}

		u := gptest.NewUnitTester(t)
		assert.NotNil(t, u)

		cfg := New()
		assert.True(t, cfg.IsSet("core.showsafecontent"))
		assert.False(t, cfg.IsSet("show.safecontent"))
	})

	t.Run("migrate submount options", func(t *testing.T) {
		u := gptest.NewUnitTester(t)
		assert.NotNil(t, u)
		// we create a submount store
		require.NoError(t, u.InitStore("submount"))

		cfg := New()
		// we add it as a mount path to our global config
		require.NoError(t, cfg.SetMountPath("submount", u.StoreDir("submount")))
		// we reload the config so the submount config is created and loaded
		cfg = New()

		// this will write to the local mount config
		require.NoError(t, cfg.Set("submount", "core.showsafecontent", "true"))
		assert.Equal(t, "true", cfg.GetM("submount", "core.showsafecontent"))
		assert.Empty(t, cfg.Get("core.showsafecontent"))
		assert.Empty(t, cfg.GetM("submount", "show.safecontent"))

		t.Setenv("GOPASS_CONFIG_NO_MIGRATE", "")
		// we test the migration path
		cfg = New()
		assert.False(t, cfg.IsSet("core.showsafecontent"))
		assert.False(t, cfg.IsSet("show.safecontent"))
		assert.Empty(t, cfg.GetM("submount", "core.showsafecontent"))
		assert.Equal(t, "true", cfg.GetM("submount", "show.safecontent"))
	})
}
