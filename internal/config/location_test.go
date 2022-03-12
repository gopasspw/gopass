package config

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPwStoreDirNoEnv(t *testing.T) { //nolint:paralleltest
	if runtime.GOOS != "windows" {
		t.Setenv("GOPASS_HOMEDIR", "/tmp")
	}

	for in, out := range map[string]string{
		"":                          filepath.Join(Homedir(), ".local", "share", "gopass", "stores", "root"),
		"work":                      filepath.Join(Homedir(), ".local", "share", "gopass", "stores", "work"),
		filepath.Join("foo", "bar"): filepath.Join(Homedir(), ".local", "share", "gopass", "stores", "foo-bar"),
	} {
		assert.Equal(t, out, PwStoreDir(in), in)
	}
}

func TestDirectory(t *testing.T) {
	t.Parallel()

	loc := configLocation()
	dir := filepath.Dir(loc)
	assert.Equal(t, dir, Directory())
}
