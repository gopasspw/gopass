package sub

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/justwatchcom/gopass/fsutil"
	"github.com/justwatchcom/gopass/store"
	"github.com/pkg/errors"
)

// Copy will copy one entry to another location. Multi-store copies are
// supported. Each entry has to be decoded and encoded for the destination
// to make sure it's encrypted for the right set of recipients.
func (s *Store) Copy(ctx context.Context, from, to string) error {
	// recursive copy?
	if s.IsDir(from) {
		if s.Exists(to) {
			return errors.Errorf("Can not copy dir to file")
		}
		sf, err := s.List("")
		if err != nil {
			return errors.Wrapf(err, "failed to list store")
		}
		destPrefix := to
		if s.IsDir(to) {
			destPrefix = filepath.Join(to, filepath.Base(from))
		}
		for _, e := range sf {
			if !strings.HasPrefix(e, strings.TrimSuffix(from, "/")+"/") {
				continue
			}
			et := filepath.Join(destPrefix, strings.TrimPrefix(e, from))
			if err := s.Copy(ctx, e, et); err != nil {
				fmt.Println(err)
			}
		}
		return nil
	}

	content, err := s.Get(ctx, from)
	if err != nil {
		return errors.Wrapf(err, "failed to get '%s' from store", from)
	}
	if err := s.Set(ctx, to, content, fmt.Sprintf("Copied from %s to %s", from, to)); err != nil {
		return errors.Wrapf(err, "failed to save '%s' to store", to)
	}
	return nil
}

// Move will move one entry from one location to another. Cross-store moves are
// supported. Moving an entry will decode it from the old location, encode it
// for the destination store with the right set of recipients and remove it
// from the old location afterwards.
func (s *Store) Move(ctx context.Context, from, to string) error {
	// recursive move?
	if s.IsDir(from) {
		if s.Exists(to) {
			return errors.Errorf("Can not move dir to file")
		}
		sf, err := s.List("")
		if err != nil {
			return errors.Wrapf(err, "failed to list store")
		}
		destPrefix := to
		if s.IsDir(to) {
			destPrefix = filepath.Join(to, filepath.Base(from))
		}
		for _, e := range sf {
			if !strings.HasPrefix(e, strings.TrimSuffix(from, "/")+"/") {
				continue
			}
			et := filepath.Join(destPrefix, strings.TrimPrefix(e, from))
			if err := s.Move(ctx, e, et); err != nil {
				fmt.Println(err)
			}
		}
		return nil
	}

	content, err := s.Get(ctx, from)
	if err != nil {
		return errors.Wrapf(err, "failed to decrypt '%s'", from)
	}
	if err := s.Set(ctx, to, content, fmt.Sprintf("Moved from %s to %s", from, to)); err != nil {
		return errors.Wrapf(err, "failed to write '%s'", to)
	}
	if err := s.Delete(ctx, from); err != nil {
		return errors.Wrapf(err, "failed to delete '%s'", from)
	}
	return nil
}

// Delete will remove an single entry from the store
func (s *Store) Delete(ctx context.Context, name string) error {
	return s.delete(ctx, name, false)
}

// Prune will remove a subtree from the Store
func (s *Store) Prune(ctx context.Context, tree string) error {
	return s.delete(ctx, tree, true)
}

// delete will either delete one file or an directory tree depending on the
// RemoveFunc given. Use nil or os.Remove for the single-file mode and
// os.RemoveAll for the recursive mode.
func (s *Store) delete(ctx context.Context, name string, recurse bool) error {
	path := s.passfile(name)
	rf := os.Remove

	if !recurse && !fsutil.IsFile(path) {
		return store.ErrNotFound
	}

	if recurse && !fsutil.IsFile(path) {
		path = filepath.Join(s.path, name)
		rf = os.RemoveAll
		if !fsutil.IsDir(path) {
			return store.ErrNotFound
		}
	}

	if err := rf(path); err != nil {
		return errors.Errorf("Failed to remove secret: %v", err)
	}

	if err := s.gitAdd(ctx, path); err != nil {
		if errors.Cause(err) == store.ErrGitNotInit {
			return nil
		}
		return errors.Wrapf(err, "failed to add '%s' to git", path)
	}
	if err := s.gitCommit(ctx, fmt.Sprintf("Remove %s from store.", name)); err != nil {
		if errors.Cause(err) == store.ErrGitNotInit {
			return nil
		}
		return errors.Wrapf(err, "failed to commit changes to git")
	}

	if s.autoSync {
		if err := s.gitPush(ctx, "", ""); err != nil {
			if errors.Cause(err) == store.ErrGitNotInit || errors.Cause(err) == store.ErrGitNoRemote {
				return nil
			}
			return errors.Wrapf(err, "failed to push change to git remote")
		}
	}

	return nil
}
