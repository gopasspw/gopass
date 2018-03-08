package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/justwatchcom/gopass/backend"
	"github.com/stretchr/testify/assert"
)

func TestHomedir(t *testing.T) {
	assert.NotEqual(t, Homedir(), "")
}

func TestNewConfig(t *testing.T) {
	assert.NoError(t, os.Setenv("GOPASS_CONFIG", filepath.Join(os.TempDir(), ".gopass.yml")))

	cfg := New()
	assert.Equal(t, false, cfg.Root.AskForMore)
	assert.NoError(t, cfg.checkDefaults())
	assert.Equal(t, backend.GPGCLI, cfg.Root.Path.Crypto)
	assert.Equal(t, backend.GitCLI, cfg.Root.Path.Sync)
	assert.Equal(t, backend.FS, cfg.Root.Path.Store)
	assert.Equal(t, "Config[Root:StoreConfig[AskForMore:false,AutoImport:true,AutoSync:true,ClipTimeout:45,NoColor:false,NoConfirm:false,NoPager:false,Path:gpgcli-gitcli-fs+file:,SafeContent:false,UseSymbols:false],Mounts(),Version:]", cfg.String())

	cfg = nil
	assert.Error(t, cfg.checkDefaults())

	cfg = &Config{
		Mounts: make(map[string]*StoreConfig, 2),
	}
	cfg.Mounts["foo"] = &StoreConfig{}
	cfg.Mounts["bar"] = &StoreConfig{}
	assert.NoError(t, cfg.checkDefaults())
	assert.Equal(t, "Config[Root:StoreConfig[AskForMore:false,AutoImport:false,AutoSync:false,ClipTimeout:0,NoColor:false,NoConfirm:false,NoPager:false,Path:gpgcli-gitcli-fs+file:,SafeContent:false,UseSymbols:false],Mounts(bar=>StoreConfig[AskForMore:false,AutoImport:false,AutoSync:false,ClipTimeout:0,NoColor:false,NoConfirm:false,NoPager:false,Path:gpgcli-gitcli-fs+file:,SafeContent:false,UseSymbols:false]foo=>StoreConfig[AskForMore:false,AutoImport:false,AutoSync:false,ClipTimeout:0,NoColor:false,NoConfirm:false,NoPager:false,Path:gpgcli-gitcli-fs+file:,SafeContent:false,UseSymbols:false]),Version:]", cfg.String())
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
