package fs

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/gopasspw/gopass/pkg/debug"
)

func walkSymlinks(path string, walkFn filepath.WalkFunc) error {
	w := &walker{
		seen: map[string]bool{},
	}

	return w.walk(path, path, walkFn)
}

type walker struct {
	seen map[string]bool
}

func (w *walker) walk(filename, linkDir string, walkFn filepath.WalkFunc) error {
	sWalkFn := func(path string, info fs.FileInfo, err error) error {
		fname, err := filepath.Rel(filename, path)
		if err != nil {
			return err
		}
		path = filepath.Join(linkDir, fname)

		if info.Mode()&fs.ModeSymlink == fs.ModeSymlink {
			destPath, err := filepath.EvalSymlinks(path)
			if err != nil {
				return err
			}

			// avoid loops
			if w.seen[destPath] {
				debug.Log("Symlink loop detected at %s!", destPath)

				return nil
			}
			w.seen[destPath] = true

			destInfo, err := os.Lstat(destPath)
			if err != nil {
				return walkFn(path, destInfo, err)
			}

			if destInfo.IsDir() {
				return w.walk(destPath, path, walkFn)
			}
		}

		return walkFn(path, info, err)
	}

	return filepath.Walk(filename, sWalkFn)
}
