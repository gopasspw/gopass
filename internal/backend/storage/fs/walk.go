package fs

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/gopasspw/gopass/pkg/debug"
)

func walkSymlinks(root string, walkFn filepath.WalkFunc) error {
	return walk(root, root, root, walkFn)
}

func walk(root, filename, linkDir string, walkFn filepath.WalkFunc) error {
	sWalkFn := func(path string, info fs.FileInfo, _ error) error {
		fname, err := filepath.Rel(filename, path)
		if err != nil {
			return err
		}
		path = filepath.Join(linkDir, fname)

		// handle non-symlinks
		if info.Mode()&fs.ModeSymlink != fs.ModeSymlink {
			return walkFn(path, info, err)
		}

		// handle symlinks
		destPath, err := filepath.EvalSymlinks(path)
		if err != nil {
			return err
		}

		// Validate that the symlink target stays within the store root.
		// A target outside the root could expose arbitrary filesystem
		// structure (information disclosure) or cause a DoS via large
		// directory trees.  Skip silently rather than erroring so that
		// a single stray symlink does not abort the entire walk.
		if destPath != root && !strings.HasPrefix(destPath, root+string(filepath.Separator)) {
			debug.Log("skipping symlink %q: target %q escapes store root", path, destPath)

			return nil
		}

		destInfo, err := os.Lstat(destPath)
		if err != nil {
			return walkFn(path, destInfo, err)
		}

		if destInfo.IsDir() {
			return walk(root, destPath, path, walkFn)
		}

		return walkFn(path, info, err)
	}

	return filepath.Walk(filename, sWalkFn)
}
