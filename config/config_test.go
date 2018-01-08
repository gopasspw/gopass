package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHomedir(t *testing.T) {
	assert.NotEqual(t, Homedir(), "")
}

func TestNewConfig(t *testing.T) {
	assert.NoError(t, os.Setenv("GOPASS_CONFIG", filepath.Join(os.TempDir(), ".gopass.yml")))

	cfg := New()
	assert.Equal(t, false, cfg.Root.AskForMore)
}

func TestSetConfigValue(t *testing.T) {
	assert.NoError(t, os.Setenv("GOPASS_CONFIG", filepath.Join(os.TempDir(), ".gopass.yml")))

	cfg := New()
	assert.NoError(t, cfg.SetConfigValue("", "autosync", "false"))
	assert.NoError(t, cfg.SetConfigValue("", "askformore", "true"))
	assert.NoError(t, cfg.SetConfigValue("", "cliptimeout", "900"))
	assert.NoError(t, cfg.SetConfigValue("", "path", "/tmp"))
	assert.Error(t, cfg.SetConfigValue("", "askformore", "yo"))

	cfg.Mounts["foo"] = &StoreConfig{}
	assert.NoError(t, cfg.SetConfigValue("foo", "autosync", "true"))
	assert.NoError(t, cfg.SetConfigValue("foo", "askformore", "true"))
	assert.NoError(t, cfg.SetConfigValue("foo", "cliptimeout", "900"))
	assert.Error(t, cfg.SetConfigValue("foo", "askformore", "yo"))
}
