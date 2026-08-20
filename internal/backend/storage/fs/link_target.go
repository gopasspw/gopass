package fs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// LinkTarget returns the relative target of a symlinked secret.
func (s *Store) LinkTarget(_ context.Context, name string) (string, bool, error) {
	if runtime.GOOS == "windows" {
		name = filepath.FromSlash(name)
	}

	path, err := s.safePath(name)
	if err != nil {
		return "", false, err
	}

	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}

		return "", false, err
	}

	if info.Mode()&os.ModeSymlink != os.ModeSymlink {
		return "", false, nil
	}

	target, err := os.Readlink(path)
	if err != nil {
		return "", true, fmt.Errorf("failed to read symlink target for %q: %w", name, err)
	}

	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(path), target)
	}

	target, err = filepath.Rel(s.path, target)
	if err != nil {
		return "", true, fmt.Errorf("failed to get relative symlink target for %q: %w", name, err)
	}

	return filepath.ToSlash(filepath.Clean(target)), true, nil
}
