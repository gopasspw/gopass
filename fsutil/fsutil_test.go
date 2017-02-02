package fsutil

import (
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"testing"
)

func TestCleanPath(t *testing.T) {
	m := map[string]string{
		"/home/user/../bob/.password-store": "/home/bob/.password-store",
		"/home/user//.password-store":       "/home/user/.password-store",
	}
	usr, err := user.Current()
	if err == nil {
		m["~/.password-store"] = usr.HomeDir + "/.password-store"
	}
	for in, out := range m {
		got := CleanPath(in)
		if out != got {
			t.Errorf("Mismatch for %s: %s != %s", in, got, out)
		}
	}
}

func TestIsDir(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "gopass-")
	if err != nil {
		t.Fatalf("Failed to create tempdir: %s", err)
	}
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()
	fn := filepath.Join(tempdir, "foo")
	if err := ioutil.WriteFile(fn, []byte("bar"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %s", err)
	}
	if !IsDir(tempdir) {
		t.Errorf("Should be a dir: %s", tempdir)
	}
	if IsDir(fn) {
		t.Errorf("Should be not dir: %s", fn)
	}
}
func TestIsFile(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "gopass-")
	if err != nil {
		t.Fatalf("Failed to create tempdir: %s", err)
	}
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()
	fn := filepath.Join(tempdir, "foo")
	if err := ioutil.WriteFile(fn, []byte("bar"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %s", err)
	}
	if IsFile(tempdir) {
		t.Errorf("Should be a dir: %s", tempdir)
	}
	if !IsFile(fn) {
		t.Errorf("Should be not dir: %s", fn)
	}
}
func TestTempdir(t *testing.T) {
	tempdir, err := ioutil.TempDir(Tempdir(), "gopass-")
	if err != nil {
		t.Fatalf("Failed to create tempdir: %s", err)
	}
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()
}
