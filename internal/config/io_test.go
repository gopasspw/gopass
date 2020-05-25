package config

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigs(t *testing.T) {
	for _, tc := range []struct {
		name string
		cfg  string
	}{
		{
			name: "1.4.0",
			cfg: `root:
  askformore: false
  autoimport: false
  autosync: false
  cliptimeout: 45
  noconfirm: false
  nopager: false
  path: /home/johndoe/.password-store
  safecontent: false
mounts:
  foo/sub:
    askformore: false
    autoimport: false
    autosync: false
    cliptimeout: 45
    noconfirm: false
    nopager: false
    path: /home/johndoe/.password-store-foo-sub
    safecontent: false
  work:
    askformore: false
    autoimport: false
    autosync: false
    cliptimeout: 45
    noconfirm: false
    nopager: false
    path: /home/johndoe/.password-store-work
    safecontent: false
version: 1.4.0`,
		}, {
			name: "1.3.0",
			cfg: `askformore: false
autoimport: true
autosync: false
cliptimeout: 45
mounts:
  dev: /Users/johndoe/.password-store-dev
  ops: /Users/johndoe/.password-store-ops
  personal: /Users/johndoe/secrets
  teststore: /Users/johndoe/tmp/teststore
noconfirm: false
path: /home/foo/.password-store
safecontent: true
version: "1.3.0"`,
		}, {
			name: "1.2.0",
			cfg: `alwaystrust: true
askformore: false
autoimport: true
autopull: true
autopush: true
cliptimeout: 45
debug: false
loadkeys: true
mounts:
  dev: /Users/johndoe/.password-store-dev
  ops: /Users/johndoe/.password-store-ops
  personal: /Users/johndoe/secrets
  teststore: /Users/johndoe/tmp/teststore
nocolor: false
noconfirm: false
path: /home/foo/.password-store
persistkeys: true
safecontent: true
version: "1.2.0"`,
		}, {
			name: "1.1.0",
			cfg: `alwaystrust: false
autoimport: false
autopull: true
autopush: true
cliptimeout: 45
loadkeys: false
mounts:
  dev: /home/johndoe/.password-store-dev
  ops: /home/johndoe/.password-store-ops
  personal: /home/johndoe/secrets
  teststore: /home/johndoe/tmp/teststore
nocolor: false
noconfirm: false
path: /home/johndoe/.password-store
persistkeys: true
safecontent: false
version: 1.1.0`,
		}, {
			name: "1.0.0",
			cfg: `alwaystrust: false
autoimport: false
autopull: true
autopush: false
cliptimeout: 45
loadkeys: false
mounts:
  dev: /Users/johndoe/.password-store-dev
  ops: /Users/johndoe/.password-store-ops
  personal: /Users/johndoe/secrets
  teststore: /Users/johndoe/tmp/teststore
noconfirm: false
path: /home/foo/.password-store
persistkeys: false
version: "1.0.0"`,
		},
	} {
		t.Logf("Loading config %s ...", tc.name)
		if _, err := decode([]byte(tc.cfg)); err != nil {
			t.Errorf("Giving up. Failed to load config %s: %s\n%s", tc.name, err, tc.cfg)
			continue
		}
		t.Logf("Success: %s", tc.name)
	}
}

const testConfig = `root:
  askformore: true
  autoimport: true
  autosync: true
  cliptimeout: 5
  noconfirm: true
  nopager: true
  path: /home/johndoe/.password-store
  safecontent: true
mounts:
  foo/sub:
    askformore: false
    autoimport: false
    autosync: false
    cliptimeout: 45
    noconfirm: false
    nopager: false
    path: /home/johndoe/.password-store-foo-sub
    safecontent: false
  work:
    askformore: false
    autoimport: false
    autosync: false
    cliptimeout: 45
    noconfirm: false
    nopager: false
    path: /home/johndoe/.password-store-work
    safecontent: false
version: 1.4.0`

func TestLoad(t *testing.T) {
	td := os.TempDir()
	gcfg := filepath.Join(td, ".gopass.yml")
	_ = os.Remove(gcfg)
	assert.NoError(t, os.Setenv("GOPASS_CONFIG", gcfg))
	assert.NoError(t, os.Setenv("GOPASS_HOMEDIR", td))

	require.NoError(t, ioutil.WriteFile(gcfg, []byte(testConfig), 0600))

	cfg := Load()
	assert.Equal(t, true, cfg.SafeContent)
}

func TestLoadError(t *testing.T) {
	gcfg := filepath.Join(os.TempDir(), ".gopass-err.yml")
	assert.NoError(t, os.Setenv("GOPASS_CONFIG", gcfg))

	_ = os.Remove(gcfg)

	capture(t, func() error {
		_, err := load(gcfg)
		if err == nil {
			return fmt.Errorf("should fail")
		}
		return nil
	})

	_ = os.Remove(gcfg)
	cfg, err := load(gcfg)
	assert.Error(t, err)

	gcfg = filepath.Join(os.TempDir(), "foo", ".gopass.yml")
	assert.NoError(t, os.Setenv("GOPASS_CONFIG", gcfg))
	assert.NoError(t, cfg.Save())
}

func TestDecodeError(t *testing.T) {
	gcfg := filepath.Join(os.TempDir(), ".gopass-err2.yml")
	assert.NoError(t, os.Setenv("GOPASS_CONFIG", gcfg))

	_ = os.Remove(gcfg)
	require.NoError(t, ioutil.WriteFile(gcfg, []byte(testConfig+"\nfoobar: zab\n"), 0600))

	capture(t, func() error {
		_, err := load(gcfg)
		if err == nil {
			return fmt.Errorf("should fail")
		}
		return nil
	})
}

func capture(t *testing.T, fn func() error) string {
	t.Helper()
	old := os.Stdout

	oldcol := color.NoColor
	color.NoColor = true

	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	done := make(chan string)
	go func() {
		buf := &bytes.Buffer{}
		_, _ = io.Copy(buf, r)
		done <- buf.String()
	}()

	err = fn()
	// back to normal
	_ = w.Close()
	os.Stdout = old
	color.NoColor = oldcol
	assert.NoError(t, err)
	out := <-done
	return strings.TrimSpace(out)
}
