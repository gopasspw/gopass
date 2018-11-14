package fsutil

import (
	"crypto/rand"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCleanFilename(t *testing.T) {
	m := map[string]string{
		`"§$%&aÜÄ*&b%§"'Ä"c%$"'"`: "a____b______c",
	}
	for k, v := range m {
		out := CleanFilename(k)
		t.Logf("%s -> %s / %s", k, v, out)
		if out != v {
			t.Errorf("'%s' != '%s'", out, v)
		}
	}
}

func TestCleanPath(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	m := map[string]string{
		".":                                 "",
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
		assert.NoError(t, err)
		assert.Equal(t, absOut, got)
	}
}

func TestIsDir(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	fn := filepath.Join(tempdir, "foo")
	assert.NoError(t, ioutil.WriteFile(fn, []byte("bar"), 0644))
	assert.Equal(t, true, IsDir(tempdir))
	assert.Equal(t, false, IsDir(fn))
	assert.Equal(t, false, IsDir(filepath.Join(tempdir, "non-existing")))
}

func TestIsFile(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	fn := filepath.Join(tempdir, "foo")
	assert.NoError(t, ioutil.WriteFile(fn, []byte("bar"), 0644))
	assert.Equal(t, false, IsFile(tempdir))
	assert.Equal(t, true, IsFile(fn))
}

func TestShred(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	fn := filepath.Join(tempdir, "file")
	// test successful shread
	fh, err := os.OpenFile(fn, os.O_WRONLY|os.O_CREATE, 0644)
	assert.NoError(t, err)

	buf := make([]byte, 1024)
	for i := 0; i < 10*1024; i++ {
		_, _ = rand.Read(buf)
		_, _ = fh.Write(buf)
	}
	_ = fh.Close()
	assert.NoError(t, Shred(fn, 8))
	assert.Equal(t, false, IsFile(fn))

	// test failed
	fh, err = os.OpenFile(fn, os.O_WRONLY|os.O_CREATE, 0400)
	assert.NoError(t, err)

	buf = make([]byte, 1024)
	for i := 0; i < 10*1024; i++ {
		_, _ = rand.Read(buf)
		_, _ = fh.Write(buf)
	}
	_ = fh.Close()
	assert.Error(t, Shred(fn, 8))
	assert.Equal(t, true, IsFile(fn))
}

func TestIsEmptyDir(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	fn := filepath.Join(tempdir, "foo", "bar", "baz", "zab")
	assert.NoError(t, os.MkdirAll(fn, 0755))

	isEmpty, err := IsEmptyDir(tempdir)
	assert.NoError(t, err)
	assert.Equal(t, true, isEmpty)

	fn = filepath.Join(fn, ".config.yml")
	assert.NoError(t, ioutil.WriteFile(fn, []byte("foo"), 0644))

	isEmpty, err = IsEmptyDir(tempdir)
	assert.NoError(t, err)
	assert.Equal(t, false, isEmpty)
}
