package fs

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWalkTooLong(t *testing.T) {
	t.Parallel()
	// Walking a path with a symlink loop should fail.

	td := t.TempDir()
	storeDir := filepath.Join(td, "store")
	fn := filepath.Join(storeDir, "real", "file.txt")
	require.NoError(t, os.MkdirAll(filepath.Dir(fn), 0o700))
	require.NoError(t, os.WriteFile(fn, []byte("foobar"), 0o600))

	ptr := filepath.Join(storeDir, "path", "via", "link")

	require.NoError(t, os.MkdirAll(filepath.Dir(ptr), 0o700))

	require.NoError(t, os.Symlink(filepath.Join(storeDir, "path"), filepath.Join(storeDir, "path", "via", "loop")))

	// test the walkFunc
	require.Error(t, walkSymlinks(storeDir, func(path string, info fs.FileInfo, err error) error {
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
	require.NoError(t, os.MkdirAll(filepath.Dir(fn), 0o700))
	require.NoError(t, os.WriteFile(fn, []byte("foobar"), 0o600))

	ptr1 := filepath.Join(storeDir, "path", "via", "one", "link")
	ptr2 := filepath.Join(storeDir, "another", "path", "to", "this", "file")

	require.NoError(t, os.MkdirAll(filepath.Dir(ptr1), 0o700))
	require.NoError(t, os.MkdirAll(filepath.Dir(ptr2), 0o700))

	require.NoError(t, os.Symlink(fn, ptr1))
	require.NoError(t, os.Symlink(fn, ptr2))

	// test the walkFunc
	seen := map[string]bool{}
	want := map[string]bool{
		"another/path/to/this/file": true,
		"path/via/one/link":         true,
		"real/file.txt":             true,
	}

	require.NoError(t, walkSymlinks(storeDir, func(path string, info fs.FileInfo, err error) error {
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

// TestWalkEscapeSymlink verifies that a directory symlink pointing outside the
// store root is silently skipped and does NOT cause files outside the store to
// appear in the walk results.
func TestWalkEscapeSymlink(t *testing.T) {
	t.Parallel()

	td := t.TempDir()

	// A directory that lives entirely outside the store.
	outsideDir := filepath.Join(td, "outside")
	secretFile := filepath.Join(outsideDir, "secret.txt")
	require.NoError(t, os.MkdirAll(outsideDir, 0o700))
	require.NoError(t, os.WriteFile(secretFile, []byte("outside-secret"), 0o600))

	// The store itself contains one legitimate file.
	storeDir := filepath.Join(td, "store")
	storeFile := filepath.Join(storeDir, "legit.age")
	require.NoError(t, os.MkdirAll(storeDir, 0o700))
	require.NoError(t, os.WriteFile(storeFile, []byte("encrypted"), 0o600))

	// A symlink inside the store that points to the outside directory.
	escapeLink := filepath.Join(storeDir, "escape")
	require.NoError(t, os.Symlink(outsideDir, escapeLink))

	seen := map[string]bool{}
	require.NoError(t, walkSymlinks(storeDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rPath := strings.TrimPrefix(path, storeDir+string(filepath.Separator))
		seen[filepath.ToSlash(rPath)] = true

		return nil
	}))

	// Only the legitimate in-store file should be seen.
	assert.Equal(t, map[string]bool{"legit.age": true}, seen,
		"files outside the store must not appear in walk results")
}
