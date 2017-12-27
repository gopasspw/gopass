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
)

func TestConfigs(t *testing.T) {
	for _, cfg := range []string{
		`root:
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
		`askformore: false
autoimport: true
autosync: false
cliptimeout: 45
mounts:
  dev: /Users/johndoe/.password-store-dev
  ops: /Users/johndoe/.password-store-ops
  personal: /Users/johndoe/secrets
  teststore: /Users/johndoe/tmp/teststore
noconfirm: false
path: /home/tex/.password-store
safecontent: true
version: "1.3.0"`,
		`alwaystrust: true
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
path: /home/tex/.password-store
persistkeys: true
safecontent: true
version: "1.2.0"`,
		`alwaystrust: false
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
		`alwaystrust: false
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
path: /home/tex/.password-store
persistkeys: false
version: "1.0.0"`,
	} {
		if _, err := decode([]byte(cfg)); err != nil {
			t.Errorf("Failed to load config: %s\n%s", err, cfg)
		}
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
	gcfg := filepath.Join(os.TempDir(), ".gopass.yml")
	if err := os.Setenv("GOPASS_CONFIG", gcfg); err != nil {
		t.Fatalf("Failed to set GOPASS_CONFIG: %s", err)
	}

	if err := ioutil.WriteFile(gcfg, []byte(testConfig), 0600); err != nil {
		t.Fatalf("Failed to write config %s: %s", gcfg, err)
	}

	cfg := Load()
	if !cfg.Root.SafeContent {
		t.Errorf("SafeContent should be true")
	}
}

func TestLoadError(t *testing.T) {
	gcfg := filepath.Join(os.TempDir(), ".gopass-err.yml")
	if err := os.Setenv("GOPASS_CONFIG", gcfg); err != nil {
		t.Fatalf("Failed to set GOPASS_CONFIG: %s", err)
	}

	_ = os.Remove(gcfg)
	if err := ioutil.WriteFile(gcfg, []byte(testConfig), 0000); err != nil {
		t.Fatalf("Failed to write config %s: %s", gcfg, err)
	}

	capture(t, func() error {
		_, err := load(gcfg)
		if err == nil {
			return fmt.Errorf("Should fail")
		}
		return nil
	})

	_ = os.Remove(gcfg)
	cfg, err := load(gcfg)
	if err == nil {
		t.Errorf("Should fail")
	}
	gcfg = filepath.Join(os.TempDir(), "foo", ".gopass.yml")
	if err := os.Setenv("GOPASS_CONFIG", gcfg); err != nil {
		t.Fatalf("Failed to set GOPASS_CONFIG: %s", err)
	}
	if err := cfg.Save(); err != nil {
		t.Errorf("Error: %s", err)
	}
}

func TestDecodeError(t *testing.T) {
	gcfg := filepath.Join(os.TempDir(), ".gopass-err2.yml")
	if err := os.Setenv("GOPASS_CONFIG", gcfg); err != nil {
		t.Fatalf("Failed to set GOPASS_CONFIG: %s", err)
	}

	_ = os.Remove(gcfg)
	if err := ioutil.WriteFile(gcfg, []byte(testConfig+"\nfoobar: zab\n"), 0600); err != nil {
		t.Fatalf("Failed to write config %s: %s", gcfg, err)
	}

	capture(t, func() error {
		_, err := load(gcfg)
		if err == nil {
			return fmt.Errorf("Should fail")
		}
		return nil
	})
}

func capture(t *testing.T, fn func() error) string {
	t.Helper()
	old := os.Stdout

	oldcol := color.NoColor
	color.NoColor = true

	r, w, _ := os.Pipe()
	os.Stdout = w

	done := make(chan string)
	go func() {
		buf := &bytes.Buffer{}
		_, _ = io.Copy(buf, r)
		done <- buf.String()
	}()

	err := fn()
	// back to normal
	_ = w.Close()
	os.Stdout = old
	color.NoColor = oldcol
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	out := <-done
	return strings.TrimSpace(out)
}
