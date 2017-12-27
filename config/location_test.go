package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPwStoreDirNoEnv(t *testing.T) {
	for in, out := range map[string]string{
		"":     filepath.Join(Homedir(), ".password-store"),
		"work": filepath.Join(Homedir(), ".password-store-work"),
		filepath.Join("foo", "bar"): filepath.Join(Homedir(), ".password-store-foo-bar"),
	} {
		got := PwStoreDir(in)
		if got != out {
			t.Errorf("Mismatch for %s: %s != %s", in, got, out)
		}
	}
}

func TestPwStoreDir(t *testing.T) {
	gph := filepath.Join(os.TempDir(), "home")
	_ = os.Setenv("GOPASS_HOMEDIR", gph)

	if d := PwStoreDir(""); d != filepath.Join(gph, ".password-store") {
		t.Errorf("Wrong dir: %s", d)
	}
	if d := PwStoreDir("foo"); d != filepath.Join(gph, ".password-store-foo") {
		t.Errorf("Wrong dir: %s", d)
	}

	psd := filepath.Join(gph, ".password-store-test")
	_ = os.Setenv("PASSWORD_STORE_DIR", psd)

	if d := PwStoreDir(""); d != psd {
		t.Errorf("Wrong dir: %s", d)
	}
	if d := PwStoreDir("foo"); d != filepath.Join(gph, ".password-store-foo") {
		t.Errorf("Wrong dir: %s", d)
	}
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
		_ = os.Unsetenv(k)
	}
	for k, v := range evs {
		_ = os.Setenv(k, v.ev)
		loc := configLocation()
		t.Logf("%s = %s -> %s", k, v.ev, loc)
		if loc != v.loc {
			t.Errorf("'%s' != '%s'", loc, v.loc)
		}
		_ = os.Unsetenv(k)
	}
}

func TestConfigLocations(t *testing.T) {
	gpcfg := filepath.Join(os.TempDir(), "config", ".gopass.yml")
	_ = os.Setenv("GOPASS_CONFIG", gpcfg)
	xdghome := filepath.Join(os.TempDir(), "xdg")
	_ = os.Setenv("XDG_CONFIG_HOME", xdghome)
	gphome := filepath.Join(os.TempDir(), "home")
	_ = os.Setenv("GOPASS_HOMEDIR", gphome)

	locs := configLocations()
	t.Logf("Locations: %+v", locs)
	if len(locs) != 4 {
		t.Errorf("Expects 4 locations not %d", len(locs))
	}
	if locs[0] != gpcfg {
		t.Errorf("'%s' != '%s'", locs[0], gpcfg)
	}
	xdgcfg := filepath.Join(xdghome, "gopass", "config.yml")
	if locs[1] != xdgcfg {
		t.Errorf("'%s' != '%s'", locs[1], xdgcfg)
	}
	curcfg := filepath.Join(gphome, ".config", "gopass", "config.yml")
	if locs[2] != curcfg {
		t.Errorf("'%s' != '%s'", locs[2], curcfg)
	}
	oldcfg := filepath.Join(gphome, ".gopass.yml")
	if locs[3] != oldcfg {
		t.Errorf("'%s' != '%s'", locs[3], oldcfg)
	}
}
