package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPwStoreDirNoEnv(t *testing.T) {
	for in, out := range map[string]string{
		"":                          filepath.Join(Homedir(), ".password-store"),
		"work":                      filepath.Join(Homedir(), ".local", "share", "gopass", "stores", "work"),
		filepath.Join("foo", "bar"): filepath.Join(Homedir(), ".local", "share", "gopass", "stores", "foo-bar"),
	} {
		assert.Equal(t, out, PwStoreDir(in))
	}
}

func TestDirectory(t *testing.T) {
	loc := configLocation()
	dir := filepath.Dir(loc)
	assert.Equal(t, dir, Directory())
}
