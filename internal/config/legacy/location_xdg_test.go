//go:build !darwin && !windows
// +build !darwin,!windows

package legacy

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigLocations(t *testing.T) {
	gpcfg := filepath.Join(os.TempDir(), "config", ".gopass.yml")
	xdghome := filepath.Join(os.TempDir(), "xdg")
	gphome := filepath.Join(os.TempDir(), "home")

	xdgcfg := filepath.Join(xdghome, "gopass", "config.yml")
	curcfg := filepath.Join(gphome, ".config", "gopass", "config.yml")
	oldcfg := filepath.Join(gphome, ".gopass.yml")

	t.Run("GOPASS_CONFIG, GOPASS_HOMEDIR set", func(t *testing.T) {
		t.Setenv("GOPASS_CONFIG", gpcfg)
		t.Setenv("GOPASS_HOMEDIR", gphome)

		assert.Equal(t, []string{gpcfg, curcfg, curcfg, oldcfg}, ConfigLocations())
	})

	t.Run("GOPASS_CONFIG, GOPASS_HOMEDIR, XDG_CONFIG_HOME set", func(t *testing.T) {
		t.Setenv("GOPASS_CONFIG", gpcfg)
		t.Setenv("GOPASS_HOMEDIR", gphome)
		t.Setenv("XDG_CONFIG_HOME", xdghome)

		assert.Equal(t, []string{gpcfg, curcfg, curcfg, oldcfg}, ConfigLocations())
	})

	t.Run("XDG_CONFIG_HOME set only", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", xdghome)
		assert.Equal(t, xdgcfg, ConfigLocations()[0])
	})
}
