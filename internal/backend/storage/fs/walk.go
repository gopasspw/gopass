package fs

import (
	"io/fs"
	"os"
	"path/filepath"
)

func walkSymlinks(path string, walkFn filepath.WalkFunc) error {
	return walk(path, path, walkFn)
}

func walk(filename, linkDir string, walkFn filepath.WalkFunc) error {
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

		destInfo, err := os.Lstat(destPath)
		if err != nil {
			return walkFn(path, destInfo, err)
		}

		if destInfo.IsDir() {
			return walk(destPath, path, walkFn)
		}

		return walkFn(path, info, err)
	}

	return filepath.Walk(filename, sWalkFn)
}
