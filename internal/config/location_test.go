package config

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/gopasspw/gopass/pkg/appdir"
	"github.com/stretchr/testify/assert"
)

func TestPwStoreDirNoEnv(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Setenv("GOPASS_HOMEDIR", "/tmp")
	}

	baseDir := filepath.Join(appdir.UserHome(), ".local", "share", "gopass", "stores")
	if runtime.GOOS == "windows" {
		baseDir = filepath.Join(appdir.UserHome(), "AppData", "Local", "gopass", "stores")
	}

	for in, out := range map[string]string{
		"":                          filepath.Join(baseDir, "root"),
		"work":                      filepath.Join(baseDir, "work"),
		filepath.Join("foo", "bar"): filepath.Join(baseDir, "foo-bar"),
	} {
		assert.Equal(t, out, PwStoreDir(in), in, "mount "+in)
	}
}

func TestDirectory(t *testing.T) {
	t.Parallel()

	loc := configLocation()
	dir := filepath.Dir(loc)
	assert.Equal(t, dir, Directory())
}
