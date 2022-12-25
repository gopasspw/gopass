package fs

import (
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWalkTooLong(t *testing.T) {
	t.Parallel()
	// Walking a path with a symlink loop should fail.

	td := t.TempDir()
	storeDir := filepath.Join(td, "store")
	fn := filepath.Join(storeDir, "real", "file.txt")
	assert.NoError(t, os.MkdirAll(filepath.Dir(fn), 0o700))
	assert.NoError(t, ioutil.WriteFile(fn, []byte("foobar"), 0o600))

	ptr := filepath.Join(storeDir, "path", "via", "link")

	assert.NoError(t, os.MkdirAll(filepath.Dir(ptr), 0o700))

	assert.NoError(t, os.Symlink(filepath.Join(storeDir, "path"), filepath.Join(storeDir, "path", "via", "loop")))

	// test the walkFunc
	assert.Error(t, walkSymlinks(storeDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") {
			return fs.SkipDir
		}
		if info.IsDir() {
			return nil
		}
		rPath := strings.TrimPrefix(path, storeDir)
		if rPath == "" {
			return nil
		}

		return nil
	}))
}

func TestWalkSameFile(t *testing.T) {
	t.Parallel()
	// Two files visible via different link chains should both end up in the result set.

	td := t.TempDir()
	storeDir := filepath.Join(td, "store")
	fn := filepath.Join(storeDir, "real", "file.txt")
	assert.NoError(t, os.MkdirAll(filepath.Dir(fn), 0o700))
	assert.NoError(t, ioutil.WriteFile(fn, []byte("foobar"), 0o600))

	ptr1 := filepath.Join(storeDir, "path", "via", "one", "link")
	ptr2 := filepath.Join(storeDir, "another", "path", "to", "this", "file")

	assert.NoError(t, os.MkdirAll(filepath.Dir(ptr1), 0o700))
	assert.NoError(t, os.MkdirAll(filepath.Dir(ptr2), 0o700))

	assert.NoError(t, os.Symlink(fn, ptr1))
	assert.NoError(t, os.Symlink(fn, ptr2))

	// test the walkFunc
	seen := map[string]bool{}
	want := map[string]bool{
		"another/path/to/this/file": true,
		"path/via/one/link":         true,
		"real/file.txt":             true,
	}

	assert.NoError(t, walkSymlinks(storeDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") {
			return fs.SkipDir
		}
		if info.IsDir() {
			return nil
		}
		rPath := strings.TrimPrefix(path, storeDir)
		if rPath == "" {
			return nil
		}
		rPath = filepath.ToSlash(rPath) // support running this test on Windows
		rPath = strings.TrimPrefix(rPath, "/")
		seen[rPath] = true

		return nil
	}))

	assert.Equal(t, want, seen)
}
