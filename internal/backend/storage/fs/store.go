// Package fs implement a password-store compatible on disk storage layout
// with unencrypted paths.
package fs

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/fsutil"
)

// Store is a fs based store.
type Store struct {
	path    string
	version semver.Version
}

// New creates a new store.
func New(dir string) *Store {
	if d, err := filepath.EvalSymlinks(dir); err == nil {
		dir = d
	}

	return &Store{
		path:    fsutil.ExpandHomedir(dir),
		version: debug.ModuleVersion("github.com/gopasspw/gopass/internal/backend/storage/fs"),
	}
}

// safePath validates that name resolves to a path under the store root,
// returning an error if a path traversal is detected. Both lexical traversal
// (e.g. "../foo") and symlinks that point outside the store root are rejected.
// Paths that would write into the git metadata directory (.git) are rejected as
// well, so a malicious remote tree can never plant a hook or config file.
func (s *Store) safePath(name string) (string, error) {
	if err := rejectGitDir(name); err != nil {
		return "", err
	}

	resolved := filepath.Join(s.path, filepath.Clean(name))
	if resolved != s.path && !strings.HasPrefix(resolved, s.path+string(filepath.Separator)) {
		return "", fmt.Errorf("path traversal detected: %q escapes store root", name)
	}

	// The lexical check above does not resolve symlinks. A symlink inside the
	// store whose target lives outside the root would pass it, so resolve the
	// path and verify the real location stays within the store as well.
	if err := s.checkSymlinkEscape(resolved); err != nil {
		return "", err
	}

	return resolved, nil
}

// rejectGitDir returns an error if any component of name refers to the git
// metadata directory (.git). This is a defence-in-depth guard: the store must
// never write into .git, otherwise a crafted remote tree could plant a hook or
// rewrite the repository config.
func rejectGitDir(name string) error {
	clean := path.Clean(filepath.ToSlash(name))
	clean = strings.TrimPrefix(clean, "./")
	for _, part := range strings.Split(clean, "/") {
		if IsGitDirName(part) {
			return fmt.Errorf("refusing to write to git metadata path %q", name)
		}
	}

	return nil
}

// IsGitDirName reports whether a single path component names the git metadata
// directory. The comparison is case-insensitive and also covers the Windows
// aliases that resolve to .git: trailing dots/spaces, the 8.3 short name
// (GIT~1) and NTFS alternate data streams (.git::$INDEX_ALLOCATION). Callers
// must pass the slashed form: filepath.Clean would turn the slashes back into
// backslashes on Windows, so the ".git/" prefix would never match.
func IsGitDirName(name string) bool {
	if i := strings.IndexByte(name, ':'); i >= 0 {
		name = name[:i]
	}
	name = strings.TrimRight(name, ". ")

	return strings.EqualFold(name, ".git") || strings.EqualFold(name, "git~1")
}

// checkSymlinkLeadingPath refuses to write through a symlinked path component.
// safePath resolves the longest existing prefix and rejects escapes, but a
// symlink that points inside the store (e.g. evil -> .git/config) passes that
// check; writing through it would let a remote tree plant files in .git. This
// mirrors git's has_symlink_leading_path.
func (s *Store) checkSymlinkLeadingPath(path string) error {
	rel, err := filepath.Rel(s.path, path)
	if err != nil {
		return err
	}

	parts := strings.Split(rel, string(filepath.Separator))
	cur := s.path
	for _, part := range parts[:len(parts)-1] {
		cur = filepath.Join(cur, part)

		fi, err := os.Lstat(cur)
		if err != nil {
			if os.IsNotExist(err) {
				// The remaining components do not exist yet.
				return nil
			}

			return err
		}
		if fi.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("refusing to write through symlinked path component %q", cur)
		}
	}

	return nil
}

// checkSymlinkEscape resolves the longest existing prefix of path and verifies
// that it does not escape the store root. This catches symlinked store entries
// (including intermediate directory symlinks) while still allowing paths whose
// trailing components do not exist yet, e.g. when creating a new secret.
func (s *Store) checkSymlinkEscape(path string) error {
	root, err := filepath.EvalSymlinks(s.path)
	if err != nil {
		// The store root itself cannot be resolved. Fall back to the
		// configured path so we do not reject every operation.
		root = s.path
	}

	// Walk up until we find a component that exists on disk. The trailing
	// components of a path that is about to be created do not exist yet.
	existing := path
	for {
		if _, err := os.Lstat(existing); err == nil {
			break
		}

		parent := filepath.Dir(existing)
		if parent == existing {
			// Reached the filesystem root without finding anything.
			return nil
		}

		existing = parent
	}

	realPath, err := filepath.EvalSymlinks(existing)
	if err != nil {
		return fmt.Errorf("failed to resolve path %q: %w", path, err)
	}

	if realPath != root && !strings.HasPrefix(realPath, root+string(filepath.Separator)) {
		return fmt.Errorf("path traversal detected: %q escapes store root", path)
	}

	return nil
}

// Get retrieves the named content.
func (s *Store) Get(ctx context.Context, name string) ([]byte, error) {
	if runtime.GOOS == "windows" {
		name = filepath.FromSlash(name)
	}

	path, err := s.safePath(name)
	if err != nil {
		return nil, err
	}
	debug.V(3).Log("Reading %s from %s", name, path)

	return os.ReadFile(path)
}

// Set writes the given content.
func (s *Store) Set(ctx context.Context, name string, value []byte) error {
	if runtime.GOOS == "windows" {
		name = filepath.FromSlash(name)
	}

	filename, err := s.safePath(name)
	if err != nil {
		return err
	}

	if err := s.checkSymlinkLeadingPath(filename); err != nil {
		return err
	}

	// Never write through a symlink: a remote tree could otherwise plant a
	// symlink (e.g. evil -> .git/config) and then a regular file at the same
	// path, writing through the link. Replace the link with a regular file.
	if fi, err := os.Lstat(filename); err == nil && fi.Mode()&os.ModeSymlink != 0 {
		if err := os.Remove(filename); err != nil {
			return err
		}
	}

	filedir := filepath.Dir(filename)

	if !fsutil.IsDir(filedir) {
		if err := os.MkdirAll(filedir, 0o700); err != nil {
			return err
		}
	}
	debug.V(3).Log("Writing %s to %q", name, filename)

	// if we ever try to write a secret that is identical (in ciphertext) to the secret in store,
	// we might want to act differently
	// (for instance, by not adding/committing/pushing the secret in git,
	//  or by panicking in the case of password generation)
	oldvalue, err := os.ReadFile(filename)
	if err == nil && bytes.Equal(oldvalue, value) {
		return store.ErrMeaninglessWrite
	}

	return os.WriteFile(filename, value, 0o644)
}

// Move moves the named entity to the new location.
func (s *Store) Move(ctx context.Context, from, to string, del bool) error {
	if runtime.GOOS == "windows" {
		from = filepath.FromSlash(from)
		to = filepath.FromSlash(to)
	}

	fromFn, err := s.safePath(from)
	if err != nil {
		return err
	}
	toFn, err := s.safePath(to)
	if err != nil {
		return err
	}
	toDir := filepath.Dir(toFn)

	if !fsutil.IsDir(toDir) {
		if err := os.MkdirAll(toDir, 0o700); err != nil {
			return fmt.Errorf("failed to create directory %q: %w", toDir, err)
		}
	}
	debug.V(3).Log("Copying %q (%q) to %q (%q)", from, fromFn, to, toFn)

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
	path, err := s.safePath(name)
	if err != nil {
		return err
	}
	debug.V(3).Log("Deleting %s from %s", name, path)

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

	debug.V(1).Log("removing empty parent dir: %q", parent)
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
	path, err := s.safePath(name)
	if err != nil {
		return false
	}
	found := fsutil.IsFile(path)
	debug.V(2).Log("Checking if '%s' exists at %s: %t", name, path, found)

	return found
}

// List returns a list of all entities
// e.g. foo, far/bar baz/.bang
// directory separator are normalized using `/`.
func (s *Store) List(ctx context.Context, prefix string) ([]string, error) {
	prefix = strings.TrimPrefix(prefix, "/")
	debug.V(2).Log("Listing %s/%s", s.path, prefix)

	files := make([]string, 0, 100)
	if err := walkSymlinks(s.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath := strings.TrimPrefix(path, s.path+string(filepath.Separator)) + string(filepath.Separator)
		if runtime.GOOS == "windows" {
			relPath = filepath.ToSlash(relPath)
		}
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && path != s.path && !strings.HasPrefix(prefix, relPath) && filepath.Base(path) != filepath.Base(prefix) {
			debug.V(3).Log("skipping dot dir (relPath: %s, prefix: %s)", relPath, prefix)

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
	path, err := s.safePath(name)
	if err != nil {
		debug.V(2).Log("%s: %s", name, err)

		return false
	}
	isDir := fsutil.IsDir(path)
	debug.V(2).Log("%s at %s is a directory? %t", name, path, isDir)

	return isDir
}

// Prune removes a named directory.
func (s *Store) Prune(ctx context.Context, prefix string) error {
	path, err := s.safePath(prefix)
	if err != nil {
		return err
	}
	debug.Log("Purning %s from %s", prefix, path)

	if err := os.RemoveAll(path); err != nil {
		return err
	}

	return s.removeEmptyParentDirectories(path)
}

// SetSymlink creates (or replaces) a symlink at name pointing at target. It is
// used to materialise symlinked secrets from a remote tree. Any existing file
// or link at name is removed first. Only relative targets that stay inside the
// store root and outside .git are accepted, so a crafted remote tree cannot
// plant a link that escapes the store or points into the git metadata.
func (s *Store) SetSymlink(_ context.Context, name, target string) error {
	if runtime.GOOS == "windows" {
		name = filepath.FromSlash(name)
	}

	path, err := s.safePath(name)
	if err != nil {
		return err
	}

	if err := s.validateSymlinkTarget(path, target); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	if _, err := os.Lstat(path); err == nil {
		if err := os.Remove(path); err != nil {
			return err
		}
	}

	return os.Symlink(target, path)
}

// validateSymlinkTarget rejects symlink targets that are absolute, escape the
// store root, or resolve into the git metadata directory. The target is
// resolved relative to the directory containing the link, matching how the
// operating system interprets it.
func (s *Store) validateSymlinkTarget(linkPath, target string) error {
	if target == "" {
		return fmt.Errorf("refusing empty symlink target")
	}
	if filepath.IsAbs(target) || path.IsAbs(target) {
		return fmt.Errorf("refusing absolute symlink target %q", target)
	}

	resolved := filepath.Join(filepath.Dir(linkPath), filepath.FromSlash(target))
	if resolved != s.path && !strings.HasPrefix(resolved, s.path+string(filepath.Separator)) {
		return fmt.Errorf("refusing symlink target %q escaping the store root", target)
	}

	rel, err := filepath.Rel(s.path, resolved)
	if err != nil {
		return err
	}
	if err := rejectGitDir(rel); err != nil {
		return err
	}

	return nil
}

// Name returns the name of this backend.
func (s *Store) Name() string {
	return "fs"
}

// Version returns the version of this backend.
func (s *Store) Version(context.Context) semver.Version {
	return s.version
}

// String implements fmt.Stringer.
func (s *Store) String() string {
	return fmt.Sprintf("fs(%s,path:%s)", s.version.String(), s.path)
}

// Path returns the path to this storage.
func (s *Store) Path() string {
	return s.path
}
