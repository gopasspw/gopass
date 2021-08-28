//go:build !darwin && !windows
// +build !darwin,!windows

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
	require.NoError(t, os.Setenv("GOPASS_HOMEDIR", gph))

	assert.Equal(t, filepath.Join(gph, ".local", "share", "gopass", "stores", "root"), PwStoreDir(""))
	assert.Equal(t, filepath.Join(gph, ".local", "share", "gopass", "stores", "foo"), PwStoreDir("foo"))

	psd := filepath.Join(gph, ".password-store-test")
	require.NoError(t, os.Setenv("PASSWORD_STORE_DIR", psd))

	assert.Equal(t, psd, PwStoreDir(""))
	assert.Equal(t, filepath.Join(gph, ".local", "share", "gopass", "stores", "foo"), PwStoreDir("foo"))

	// GOPASS_HOMEDIR takes precedence
	require.NoError(t, os.Setenv("XDG_DATA_HOME", filepath.Join(os.TempDir(), ".local", "foo")))
	assert.Equal(t, psd, PwStoreDir(""))
	assert.Equal(t, filepath.Join(gph, ".local", "share", "gopass", "stores", "foo"), PwStoreDir("foo"))
	assert.NoError(t, os.Unsetenv("XDG_DATA_HOME"))

	// GOPASS_HOMEDIR unset, XDG_DATA_HOME takes precedence
	require.NoError(t, os.Unsetenv("GOPASS_HOMEDIR"))
	require.NoError(t, os.Setenv("XDG_DATA_HOME", filepath.Join(os.TempDir(), ".local", "foo")))
	assert.Equal(t, psd, PwStoreDir(""))
	assert.Equal(t, filepath.Join(os.TempDir(), ".local", "foo", "gopass", "stores", "foo"), PwStoreDir("foo"))
	assert.NoError(t, os.Unsetenv("XDG_DATA_HOME"))
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
		assert.NoError(t, os.Unsetenv(k))
	}

	for k, v := range evs {
		assert.NoError(t, os.Setenv(k, v.ev))
		assert.Equal(t, v.loc, configLocation())
		assert.NoError(t, os.Unsetenv(k))
	}
}

func TestConfigLocations(t *testing.T) {
	gpcfg := filepath.Join(os.TempDir(), "config", ".gopass.yml")
	xdghome := filepath.Join(os.TempDir(), "xdg")
	gphome := filepath.Join(os.TempDir(), "home")

	xdgcfg := filepath.Join(xdghome, "gopass", "config.yml")
	curcfg := filepath.Join(gphome, ".config", "gopass", "config.yml")
	oldcfg := filepath.Join(gphome, ".gopass.yml")

	assert.NoError(t, os.Setenv("GOPASS_CONFIG", gpcfg))
	assert.NoError(t, os.Setenv("GOPASS_HOMEDIR", gphome))

	assert.Equal(t, []string{gpcfg, curcfg, curcfg, oldcfg}, configLocations())

	assert.NoError(t, os.Setenv("XDG_CONFIG_HOME", xdghome))
	assert.Equal(t, []string{gpcfg, curcfg, curcfg, oldcfg}, configLocations())

	require.NoError(t, os.Setenv("XDG_CONFIG_HOME", xdghome))
	require.NoError(t, os.Unsetenv("GOPASS_HOMEDIR"))
	require.NoError(t, os.Unsetenv("GOPASS_CONFIG"))
	assert.Equal(t, xdgcfg, configLocations()[0])
}
