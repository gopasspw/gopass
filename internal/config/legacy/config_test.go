package legacy_test

import (
	"os"
	"path/filepath"
	"testing"

	_ "github.com/gopasspw/gopass/internal/backend/crypto"
	_ "github.com/gopasspw/gopass/internal/backend/storage"
	"github.com/gopasspw/gopass/internal/config/legacy"
	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	t.Setenv("GOPASS_CONFIG", filepath.Join(os.TempDir(), ".gopass.yml"))

	cfg := legacy.New()
	cs := cfg.String()
	assert.Contains(t, cs, `&legacy.Config{AutoClip:false, AutoImport:false, ClipTimeout:45, ExportKeys:true, NoPager:false, Notifications:true,`)
	assert.Contains(t, cs, `SafeContent:false, Mounts:map[string]string{},`)

	cfg = &legacy.Config{
		Mounts: map[string]string{
			"foo": "",
			"bar": "",
		},
	}
	cs = cfg.String()
	assert.Contains(t, cs, `&legacy.Config{AutoClip:false, AutoImport:false, ClipTimeout:0, ExportKeys:false, NoPager:false, Notifications:false,`)
	assert.Contains(t, cs, `SafeContent:false, Mounts:map[string]string{"bar":"", "foo":""},`)
}

func TestSetConfigValue(t *testing.T) {
	t.Setenv("GOPASS_CONFIG", filepath.Join(os.TempDir(), ".gopass.yml"))

	cfg := legacy.New()
	assert.NoError(t, cfg.SetConfigValue("autoclip", "true"))
	assert.NoError(t, cfg.SetConfigValue("cliptimeout", "900"))
	assert.NoError(t, cfg.SetConfigValue("path", "/tmp"))
	assert.Error(t, cfg.SetConfigValue("autoclip", "yo"))
}
