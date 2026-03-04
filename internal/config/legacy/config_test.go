package legacy_test

import (
	"os"
	"path/filepath"
	"testing"

	_ "github.com/gopasspw/gopass/internal/backend/crypto"
	_ "github.com/gopasspw/gopass/internal/backend/storage"
	"github.com/gopasspw/gopass/internal/config/legacy"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	u := gptest.NewUnitTester(t)
	assert.NotNil(t, u)

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
	u := gptest.NewUnitTester(t)
	assert.NotNil(t, u)

	t.Setenv("GOPASS_CONFIG", filepath.Join(os.TempDir(), ".gopass.yml"))

	cfg := legacy.New()
	require.NoError(t, cfg.SetConfigValue("autoclip", "true"))
	require.NoError(t, cfg.SetConfigValue("cliptimeout", "900"))
	require.NoError(t, cfg.SetConfigValue("path", "/tmp"))
	require.Error(t, cfg.SetConfigValue("autoclip", "yo"))
}
