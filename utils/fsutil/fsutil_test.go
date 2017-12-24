package fsutil

import (
	"crypto/rand"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"testing"
)

func TestCleanPath(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "gopass-")
	if err != nil {
		t.Fatalf("Failed to create tempdir: %s", err)
	}
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()
	m := map[string]string{
		".": "",
		"/home/user/../bob/.password-store": "/home/bob/.password-store",
		"/home/user//.password-store":       "/home/user/.password-store",
		tempdir + "/foo.gpg":                tempdir + "/foo.gpg",
	}
	usr, err := user.Current()
	if err == nil {
		m["~/.password-store"] = usr.HomeDir + "/.password-store"
	}
	for in, out := range m {
		got := CleanPath(in)

		// filepath.Abs turns /home/bob into C:\home\bob on Windows
		absOut, err := filepath.Abs(out)
		if err != nil {
			t.Errorf("filepath.Absolute errored: %s", err)
		}
		if absOut != got {
			t.Errorf("Mismatch for %s: %s != %s", in, got, absOut)
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
	tempdir, err := ioutil.TempDir(tempdirBase(), "gopass-")
	if err != nil {
		t.Fatalf("Failed to create tempdir: %s", err)
	}
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()
}

func TestShred(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "gopass-")
	if err != nil {
		t.Fatalf("Failed to create tempdir: %s", err)
	}
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	fn := filepath.Join(tempdir, "file")
	fh, err := os.OpenFile(fn, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		t.Fatalf("Failed to open file: %s", err)
	}
	buf := make([]byte, 1024)
	for i := 0; i < 10*1024; i++ {
		_, _ = rand.Read(buf)
		_, _ = fh.Write(buf)
	}
	_ = fh.Close()
	if err := Shred(fn, 8); err != nil {
		t.Fatalf("Failed to shred the file: %s", err)
	}
	if IsFile(fn) {
		t.Errorf("Failed still exists after shreding: %s", fn)
	}
}

func TestIsEmptyDir(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "gopass-")
	if err != nil {
		t.Fatalf("Failed to create tempdir: %s", err)
	}
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	fn := filepath.Join(tempdir, "foo", "bar", "baz", "zab")
	if err := os.MkdirAll(fn, 0755); err != nil {
		t.Fatalf("failed to create dir %s: %s", fn, err)
	}

	isEmpty, err := IsEmptyDir(tempdir)
	if err != nil {
		t.Fatalf("Error: %s", err)
	}
	if !isEmpty {
		t.Errorf("Dir should be empty")
	}

	fn = filepath.Join(fn, ".config.yml")
	if err := ioutil.WriteFile(fn, []byte("foo"), 0644); err != nil {
		t.Fatalf("Failed to write file %s: %s", fn, err)
	}

	isEmpty, err = IsEmptyDir(tempdir)
	if err != nil {
		t.Fatalf("Error: %s", err)
	}
	if isEmpty {
		t.Errorf("Dir should not be empty")
	}
}
