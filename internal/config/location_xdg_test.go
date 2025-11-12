//go:build !darwin && !windows

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPwStoreDir(t *testing.T) {
	gph := filepath.Join(os.TempDir(), "home")
	t.Setenv("GOPASS_HOMEDIR", gph)

	assert.Equal(t, filepath.Join(gph, ".local", "share", "gopass", "stores", "root"), PwStoreDir(""))
	assert.Equal(t, filepath.Join(gph, ".local", "share", "gopass", "stores", "foo"), PwStoreDir("foo"))

	psd := filepath.Join(gph, ".password-store-test")
	t.Setenv("PASSWORD_STORE_DIR", psd)

	assert.Equal(t, psd, PwStoreDir(""))
	assert.Equal(t, filepath.Join(gph, ".local", "share", "gopass", "stores", "foo"), PwStoreDir("foo"))

	t.Run("GOPASS_HOMEDIR takes precedence", func(t *testing.T) {
		t.Setenv("XDG_DATA_HOME", filepath.Join(os.TempDir(), ".local", "foo"))
		assert.Equal(t, psd, PwStoreDir(""))
		assert.Equal(t, filepath.Join(gph, ".local", "share", "gopass", "stores", "foo"), PwStoreDir("foo"))
	})

	t.Run("GOPASS_HOMEDIR unset, XDG_DATA_HOME takes precedence", func(t *testing.T) {
		t.Setenv("XDG_DATA_HOME", filepath.Join(os.TempDir(), ".local", "foo"))
		require.NoError(t, os.Unsetenv("GOPASS_HOMEDIR"))
		assert.Equal(t, psd, PwStoreDir(""))
		assert.Equal(t, filepath.Join(os.TempDir(), ".local", "foo", "gopass", "stores", "foo"), PwStoreDir("foo"))
	})
}

func TestConfigLocation(t *testing.T) {
	evs := map[string]struct {
		ev  string
		loc string
	}{
		"GOPASS_CONFIG":   {ev: filepath.Join(os.TempDir(), "gopass.yml"), loc: filepath.Join(os.TempDir(), "gopass.yml")},
		"XDG_CONFIG_HOME": {ev: filepath.Join(os.TempDir(), "xdg"), loc: filepath.Join(os.TempDir(), "xdg", "gopass", "config.yml")},
		"GOPASS_HOMEDIR":  {ev: filepath.Join(os.TempDir(), "home"), loc: filepath.Join(os.TempDir(), "home", ".config", "gopass", "config.yml")},
	}

	for k := range evs {
		require.NoError(t, os.Unsetenv(k))
	}

	for k, v := range evs {
		t.Run(k, func(t *testing.T) {
			t.Setenv(k, v.ev)
			assert.Equal(t, v.loc, configLocation())
		})
	}
}
