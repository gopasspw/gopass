//go:build !windows

package appdir

import (
	"testing"

	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
)

func TestUserConfig(t *testing.T) {
	ov := gptest.UnsetVars("GOPASS_HOMEDIR", "XDG_CONFIG_HOME", "HOME")
	defer ov()

	t.Run("gopass homedir", func(t *testing.T) {
		t.Setenv("GOPASS_HOMEDIR", "/foo/bar")
		assert.Equal(t, "/foo/bar/.config/gopass", UserConfig())
	})

	t.Run("xdg_config_home", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", "/foo/baz/myconfig")
		assert.Equal(t, "/foo/baz/myconfig/gopass", UserConfig())
	})

	t.Run("default", func(t *testing.T) {
		t.Setenv("HOME", "/home/gopass")
		assert.Equal(t, "/home/gopass/.config/gopass", UserConfig())
	})
}

func TestUserCache(t *testing.T) {
	ov := gptest.UnsetVars("GOPASS_HOMEDIR", "XDG_CACHE_HOME", "HOME")
	defer ov()

	t.Run("gopass homedir", func(t *testing.T) {
		t.Setenv("GOPASS_HOMEDIR", "/foo/bar")
		assert.Equal(t, "/foo/bar/.cache/gopass", UserCache())
	})

	t.Run("xdg_cache_home", func(t *testing.T) {
		t.Setenv("XDG_CACHE_HOME", "/foo/baz/mycache")
		assert.Equal(t, "/foo/baz/mycache/gopass", UserCache())
	})

	t.Run("default", func(t *testing.T) {
		t.Setenv("HOME", "/home/gopass")
		assert.Equal(t, "/home/gopass/.cache/gopass", UserCache())
	})
}

func TestUserData(t *testing.T) {
	ov := gptest.UnsetVars("GOPASS_HOMEDIR", "XDG_DATA_HOME", "HOME")
	defer ov()

	t.Run("gopass homedir", func(t *testing.T) {
		t.Setenv("GOPASS_HOMEDIR", "/foo/bar")
		assert.Equal(t, "/foo/bar/.local/share/gopass", UserData())
	})

	t.Run("xdg_data_home", func(t *testing.T) {
		t.Setenv("XDG_DATA_HOME", "/foo/baz/mydata")
		assert.Equal(t, "/foo/baz/mydata/gopass", UserData())
	})

	t.Run("default", func(t *testing.T) {
		t.Setenv("HOME", "/home/gopass")
		assert.Equal(t, "/home/gopass/.local/share/gopass", UserData())
	})
}
