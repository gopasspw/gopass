package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPwStoreDirNoEnv(t *testing.T) {
	for in, out := range map[string]string{
		"":                          filepath.Join(Homedir(), ".password-store"),
		"work":                      filepath.Join(Homedir(), ".password-store-work"),
		filepath.Join("foo", "bar"): filepath.Join(Homedir(), ".password-store-foo-bar"),
	} {
		assert.Equal(t, out, PwStoreDir(in))
	}
}

func TestPwStoreDir(t *testing.T) {
	gph := filepath.Join(os.TempDir(), "home")
	assert.NoError(t, os.Setenv("GOPASS_HOMEDIR", gph))

	assert.Equal(t, filepath.Join(gph, ".password-store"), PwStoreDir(""))
	assert.Equal(t, filepath.Join(gph, ".password-store-foo"), PwStoreDir("foo"))

	psd := filepath.Join(gph, ".password-store-test")
	assert.NoError(t, os.Setenv("PASSWORD_STORE_DIR", psd))

	assert.Equal(t, psd, PwStoreDir(""))
	assert.Equal(t, filepath.Join(gph, ".password-store-foo"), PwStoreDir("foo"))
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
	assert.NoError(t, os.Setenv("XDG_CONFIG_HOME", xdghome))
	assert.NoError(t, os.Setenv("GOPASS_HOMEDIR", gphome))

	locs := configLocations()
	t.Logf("Locations: %+v", locs)

	assert.Equal(t, 4, len(locs))
	assert.Equal(t, gpcfg, locs[0])
	assert.Equal(t, xdgcfg, locs[1])
	assert.Equal(t, curcfg, locs[2])
	assert.Equal(t, oldcfg, locs[3])
}

func TestDirectory(t *testing.T) {
	loc := configLocation()
	dir := filepath.Dir(loc)
	assert.Equal(t, dir, Directory())
}
