package fsutil

import (
	"crypto/rand"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCleanFilename(t *testing.T) {
	t.Parallel()

	m := map[string]string{
		`"§$%&aÜÄ*&b%§"'Ä"c%$"'"`: "a____b______c",
	}
	for k, v := range m {
		out := CleanFilename(k)
		t.Logf("%s -> %s / %s", k, v, out)

		assert.Equal(t, v, out)
	}
}

func TestCleanPath(t *testing.T) {
	tempdir := t.TempDir()
	t.Setenv("GOPASS_HOMEDIR", "")

	m := map[string]string{
		".":                                 "",
		"/home/user/../bob/.password-store": "/home/bob/.password-store",
		"/home/user//.password-store":       "/home/user/.password-store",
		tempdir + "/foo.gpg":                tempdir + "/foo.gpg",
		"~/.password-store":                 "~/.password-store",
	}

	for in, out := range m {
		got := CleanPath(in)

		if strings.HasPrefix(out, "~") {
			// skip these tests on windows
			if runtime.GOOS == "windows" {
				continue
			}
			assert.Equal(t, out, got)

			continue
		}
		// filepath.Abs turns /home/bob into C:\home\bob on Windows
		absOut, err := filepath.Abs(out)
		assert.NoError(t, err)
		assert.Equal(t, absOut, got)
	}
}

func TestIsDir(t *testing.T) {
	t.Parallel()

	tempdir := t.TempDir()

	fn := filepath.Join(tempdir, "foo")
	assert.NoError(t, os.WriteFile(fn, []byte("bar"), 0o644))
	assert.Equal(t, true, IsDir(tempdir))
	assert.Equal(t, false, IsDir(fn))
	assert.Equal(t, false, IsDir(filepath.Join(tempdir, "non-existing")))
}

func TestIsFile(t *testing.T) {
	t.Parallel()

	tempdir := t.TempDir()

	fn := filepath.Join(tempdir, "foo")
	assert.NoError(t, os.WriteFile(fn, []byte("bar"), 0o644))
	assert.Equal(t, false, IsFile(tempdir))
	assert.Equal(t, true, IsFile(fn))
}

func TestShred(t *testing.T) {
	t.Parallel()

	tempdir := t.TempDir()

	fn := filepath.Join(tempdir, "file")
	// test successful shread
	fh, err := os.OpenFile(fn, os.O_WRONLY|os.O_CREATE, 0o644)
	assert.NoError(t, err)

	buf := make([]byte, 1024)
	for i := 0; i < 10*1024; i++ {
		_, _ = rand.Read(buf)
		_, _ = fh.Write(buf)
	}

	require.NoError(t, fh.Close())
	assert.NoError(t, Shred(fn, 8))
	assert.Equal(t, false, IsFile(fn))

	// test failed
	fh, err = os.OpenFile(fn, os.O_WRONLY|os.O_CREATE, 0o400)
	assert.NoError(t, err)

	buf = make([]byte, 1024)
	for i := 0; i < 10*1024; i++ {
		_, _ = rand.Read(buf)
		_, _ = fh.Write(buf)
	}

	require.NoError(t, fh.Close())
	assert.Error(t, Shred(fn, 8))
	assert.Equal(t, true, IsFile(fn))
}

func TestIsEmptyDir(t *testing.T) {
	t.Parallel()

	tempdir := t.TempDir()

	fn := filepath.Join(tempdir, "foo", "bar", "baz", "zab")
	require.NoError(t, os.MkdirAll(fn, 0o755))

	isEmpty, err := IsEmptyDir(tempdir)
	assert.NoError(t, err)
	assert.Equal(t, true, isEmpty)

	fn = filepath.Join(fn, ".config.yml")
	require.NoError(t, os.WriteFile(fn, []byte("foo"), 0o644))

	isEmpty, err = IsEmptyDir(tempdir)
	require.NoError(t, err)
	assert.Equal(t, false, isEmpty)
}

func TestCopyFile(t *testing.T) {
	t.Parallel()

	tempdir := t.TempDir()

	sfn := filepath.Join(tempdir, "foo")
	require.NoError(t, os.MkdirAll(filepath.Dir(sfn), 0o755))
	require.NoError(t, os.WriteFile(sfn, []byte("foo"), 0o644))

	dfn := filepath.Join(tempdir, "bar")

	assert.NoError(t, CopyFile(sfn, dfn))

	// try to overwrite existing file w/o write bit
	dfn = filepath.Join(tempdir, "bar2")
	require.NoError(t, os.WriteFile(dfn, []byte("foo"), 0o400))
	assert.Error(t, CopyFile(sfn, dfn))
	assert.NoError(t, CopyFileForce(sfn, dfn))
}
