package fsutil

import (
	"crypto/rand"
	"os"
	"path/filepath"
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

	home, err := os.UserHomeDir()
	if err != nil {
		home = "~"
	}

	m := map[string]string{
		".":                                 "",
		"/home/user/../bob/.password-store": "/home/bob/.password-store",
		"/home/user//.password-store":       "/home/user/.password-store",
		tempdir + "/foo.gpg":                tempdir + "/foo.gpg",
		"~/.password-store":                 home + "/.password-store",
	}

	for in, out := range m {
		got := CleanPath(in)

		if strings.HasPrefix(out, "~") {
			assert.Equal(t, out, got)

			continue
		}
		// filepath.Abs turns /home/bob into C:\home\bob on Windows
		absOut, err := filepath.Abs(out)
		require.NoError(t, err)
		assert.Equal(t, absOut, got)
	}
}

func TestCleanPathWithGopassHomedir(t *testing.T) {
	tempdir := t.TempDir()
	t.Setenv("GOPASS_HOMEDIR", tempdir)

	// ~/.local/... must expand to $GOPASS_HOMEDIR/.local/..., not $GOPASS_HOMEDIR.local/...
	assert.Equal(t, filepath.Join(tempdir, ".local", "share"), CleanPath("~/.local/share"))
	assert.Equal(t, filepath.Join(tempdir, ".password-store"), CleanPath("~/.password-store"))
}

func TestShrinkPath(t *testing.T) {
	tempdir := t.TempDir()

	t.Run("with GOPASS_HOMEDIR", func(t *testing.T) {
		t.Setenv("GOPASS_HOMEDIR", tempdir)

		assert.Equal(t, "~/.local/share/gopass", ShrinkPath(filepath.Join(tempdir, ".local", "share", "gopass")))
		assert.Equal(t, "~/.password-store", ShrinkPath(filepath.Join(tempdir, ".password-store")))
		// path outside GOPASS_HOMEDIR should be returned as-is
		assert.Equal(t, "/tmp/other", ShrinkPath("/tmp/other"))
	})

	t.Run("without GOPASS_HOMEDIR", func(t *testing.T) {
		t.Setenv("GOPASS_HOMEDIR", "")

		home, err := os.UserHomeDir()
		require.NoError(t, err)

		assert.Equal(t, "~/.local/share/gopass", ShrinkPath(filepath.Join(home, ".local", "share", "gopass")))
		// path outside home dir should be returned as-is
		assert.Equal(t, "/tmp/other", ShrinkPath("/tmp/other"))
	})
}

func TestIsDir(t *testing.T) {
	t.Parallel()

	tempdir := t.TempDir()

	fn := filepath.Join(tempdir, "foo")
	require.NoError(t, os.WriteFile(fn, []byte("bar"), 0o644))
	assert.True(t, IsDir(tempdir))
	assert.False(t, IsDir(fn))
	assert.False(t, IsDir(filepath.Join(tempdir, "non-existing")))
}

func TestIsFile(t *testing.T) {
	t.Parallel()

	tempdir := t.TempDir()

	fn := filepath.Join(tempdir, "foo")
	require.NoError(t, os.WriteFile(fn, []byte("bar"), 0o644))
	assert.False(t, IsFile(tempdir))
	assert.True(t, IsFile(fn))
}

func TestShred(t *testing.T) {
	t.Parallel()

	tempdir := t.TempDir()

	fn := filepath.Join(tempdir, "file")
	// test successful shread
	fh, err := os.OpenFile(fn, os.O_WRONLY|os.O_CREATE, 0o644)
	require.NoError(t, err)

	buf := make([]byte, 1024)
	for range 10 * 1024 {
		_, _ = rand.Read(buf)
		_, _ = fh.Write(buf)
	}

	require.NoError(t, fh.Close())
	require.NoError(t, Shred(fn, 8))
	assert.False(t, IsFile(fn))

	// test failed
	fh, err = os.OpenFile(fn, os.O_WRONLY|os.O_CREATE, 0o400)
	require.NoError(t, err)

	buf = make([]byte, 1024)
	for range 10 * 1024 {
		_, _ = rand.Read(buf)
		_, _ = fh.Write(buf)
	}

	require.NoError(t, fh.Close())
	require.Error(t, Shred(fn, 8))
	assert.True(t, IsFile(fn))
}

func TestIsEmptyDir(t *testing.T) {
	t.Parallel()

	tempdir := t.TempDir()

	fn := filepath.Join(tempdir, "foo", "bar", "baz", "zab")
	require.NoError(t, os.MkdirAll(fn, 0o755))

	isEmpty, err := IsEmptyDir(tempdir)
	require.NoError(t, err)
	assert.True(t, isEmpty)

	fn = filepath.Join(fn, ".config.yml")
	require.NoError(t, os.WriteFile(fn, []byte("foo"), 0o644))

	isEmpty, err = IsEmptyDir(tempdir)
	require.NoError(t, err)
	assert.False(t, isEmpty)
}

func TestCopyFile(t *testing.T) {
	t.Parallel()

	tempdir := t.TempDir()

	sfn := filepath.Join(tempdir, "foo")
	require.NoError(t, os.MkdirAll(filepath.Dir(sfn), 0o755))
	require.NoError(t, os.WriteFile(sfn, []byte("foo"), 0o644))

	dfn := filepath.Join(tempdir, "bar")

	require.NoError(t, CopyFile(sfn, dfn))

	// try to overwrite existing file w/o write bit
	dfn = filepath.Join(tempdir, "bar2")
	require.NoError(t, os.WriteFile(dfn, []byte("foo"), 0o400))
	require.Error(t, CopyFile(sfn, dfn))
	require.NoError(t, CopyFileForce(sfn, dfn))
}
