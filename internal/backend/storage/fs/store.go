// Package fs implement a password-store compatible on disk storage layout
// with unencrypted paths.
package fs

import (
	"context"
	"fmt"
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/fsutil"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store"
)

// Store is a fs based store.
type Store struct {
	path string
}

// New creates a new store.
func New(dir string) *Store {
	if d, err := filepath.EvalSymlinks(dir); err == nil {
		dir = d
	}

	return &Store{
		path: dir,
	}
}

// Get retrieves the named content.
func (s *Store) Get(ctx context.Context, name string) ([]byte, error) {
	if runtime.GOOS == "windows" {
		name = filepath.FromSlash(name)
	}

	path := filepath.Join(s.path, filepath.Clean(name))
	debug.Log("Reading %s from %s", name, path)

	return os.ReadFile(path)
}

// Set writes the given content.
func (s *Store) Set(ctx context.Context, name string, value []byte) error {
	if runtime.GOOS == "windows" {
		name = filepath.FromSlash(name)
	}

	filename := filepath.Join(s.path, filepath.Clean(name))
	filedir := filepath.Dir(filename)

	if !fsutil.IsDir(filedir) {
		if err := os.MkdirAll(filedir, 0o700); err != nil {
			return err
		}
	}
	debug.Log("Writing %s to %s", name, filepath.Join(s.path, name))

	oldvalue, err := os.ReadFile(filepath.Join(s.path, name))
	//out.Warningf(ctx, "%h\n%h\n%w", oldvalue, value, err)
	if err == nil && bytes.Compare(oldvalue,value)==0 {
		return store.ErrMeaninglessWrite
	}

	return os.WriteFile(filepath.Join(s.path, name), value, 0o644)
}

// Move moves the named entity to the new location.
func (s *Store) Move(ctx context.Context, from, to string, del bool) error {
	if runtime.GOOS == "windows" {
		from = filepath.FromSlash(from)
		to = filepath.FromSlash(to)
	}

	fromFn := filepath.Join(s.path, filepath.Clean(from))
	toFn := filepath.Join(s.path, filepath.Clean(to))
	toDir := filepath.Dir(toFn)

	if !fsutil.IsDir(toDir) {
		if err := os.MkdirAll(toDir, 0o700); err != nil {
			return fmt.Errorf("failed to create directory %q: %w", toDir, err)
		}
	}
	debug.Log("Copying %q (%q) to %q (%q)", from, fromFn, to, toFn)

	if del {
		if err := os.Rename(fromFn, toFn); err != nil {
			return fmt.Errorf("failed to copy %q to %q: %w", from, to, err)
		}

		return s.removeEmptyParentDirectories(fromFn)
	}

	return fsutil.CopyFile(fromFn, toFn)
}

// Delete removes the named entity.
func (s *Store) Delete(ctx context.Context, name string) error {
	if runtime.GOOS == "windows" {
		name = filepath.FromSlash(name)
	}
	path := filepath.Join(s.path, filepath.Clean(name))
	debug.Log("Deleting %s from %s", name, path)

	if err := os.Remove(path); err != nil {
		return err
	}

	return s.removeEmptyParentDirectories(path)
}

// Deletes all empty parent directories up to the store root.
func (s *Store) removeEmptyParentDirectories(path string) error {
	if runtime.GOOS == "windows" {
		path = filepath.FromSlash(path)
	}
	parent := filepath.Dir(path)

	if relpath, err := filepath.Rel(s.path, parent); err != nil {
		return err
	} else if strings.HasPrefix(relpath, ".") {
		return nil
	}

	debug.Log("removing empty parent dir: %q", parent)
	err := os.Remove(parent)
	switch {
	case err == nil:
		return s.removeEmptyParentDirectories(parent)
	case notEmptyErr(err):
		// ignore when directory is non-empty.
		return nil
	default:
		return err
	}
}

// Exists checks if the named entity exists.
func (s *Store) Exists(ctx context.Context, name string) bool {
	if runtime.GOOS == "windows" {
		name = filepath.FromSlash(name)
	}
	path := filepath.Join(s.path, filepath.Clean(name))
	found := fsutil.IsFile(path)
	debug.Log("Checking if %s exists at %s: %t", name, path, found)

	return found
}

// List returns a list of all entities
// e.g. foo, far/bar baz/.bang
// directory separator are normalized using `/`.
func (s *Store) List(ctx context.Context, prefix string) ([]string, error) {
	prefix = strings.TrimPrefix(prefix, "/")
	debug.Log("Listing %s/%s", s.path, prefix)

	files := make([]string, 0, 100)
	if err := walkSymlinks(s.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath := strings.TrimPrefix(path, s.path+string(filepath.Separator)) + string(filepath.Separator)
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && path != s.path && !strings.HasPrefix(prefix, relPath) && filepath.Base(path) != filepath.Base(prefix) {
			debug.Log("skipping dot dir (relPath: %s, prefix: %s)", relPath, prefix)

			return filepath.SkipDir
		}
		if info.IsDir() {
			return nil
		}
		if path == s.path {
			return nil
		}
		name := strings.TrimPrefix(path, s.path+string(filepath.Separator))
		if runtime.GOOS == "windows" {
			name = filepath.ToSlash(name)
		}
		if !strings.HasPrefix(name, prefix) {
			return nil
		}

		files = append(files, name)

		return nil
	}); err != nil {
		return nil, err
	}
	sort.Strings(files)

	return files, nil
}

// IsDir returns true if the named entity is a directory.
func (s *Store) IsDir(ctx context.Context, name string) bool {
	if runtime.GOOS == "windows" {
		name = filepath.FromSlash(name)
	}
	path := filepath.Join(s.path, filepath.Clean(name))
	isDir := fsutil.IsDir(path)
	debug.Log("%s at %s is a directory? %t", name, path, isDir)

	return isDir
}

// Prune removes a named directory.
func (s *Store) Prune(ctx context.Context, prefix string) error {
	path := filepath.Join(s.path, filepath.Clean(prefix))
	debug.Log("Purning %s from %s", prefix, path)

	if err := os.RemoveAll(path); err != nil {
		return err
	}

	return s.removeEmptyParentDirectories(path)
}

// Name returns the name of this backend.
func (s *Store) Name() string {
	return "fs"
}

// Version returns the version of this backend.
func (s *Store) Version(context.Context) semver.Version {
	return debug.ModuleVersion("github.com/gopasspw/gopass/internal/backend/fs")
}

// String implements fmt.Stringer.
func (s *Store) String() string {
	return fmt.Sprintf("fs(%s,path:%s)", s.Version(context.TODO()).String(), s.path)
}

// Path returns the path to this storage.
func (s *Store) Path() string {
	return s.path
}
