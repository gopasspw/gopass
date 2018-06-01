package sub

import (
	"context"
	"fmt"

	"github.com/gopasspw/gopass/pkg/store"

	"github.com/pkg/errors"
)

// Copy will copy one entry to another location. Multi-store copies are
// supported. Each entry has to be decoded and encoded for the destination
// to make sure it's encrypted for the right set of recipients.
func (s *Store) Copy(ctx context.Context, from, to string) error {
	// recursive copy?
	if s.IsDir(ctx, from) {
		return errors.Errorf("recursive operations are not supported")
	}

	content, err := s.Get(ctx, from)
	if err != nil {
		return errors.Wrapf(err, "failed to get '%s' from store", from)
	}
	if err := s.Set(WithReason(ctx, fmt.Sprintf("Copied from %s to %s", from, to)), to, content); err != nil {
		return errors.Wrapf(err, "failed to save '%s' to store", to)
	}
	return nil
}

// Move will move one entry from one location to another.
// Moving an entry will decode it from the old location, encode it
// for the destination store with the right set of recipients and remove it
// from the old location afterwards.
func (s *Store) Move(ctx context.Context, from, to string) error {
	// recursive move?
	if s.IsDir(ctx, from) {
		return errors.Errorf("recursive operations are not supported")
	}

	content, err := s.Get(ctx, from)
	if err != nil {
		return errors.Wrapf(err, "failed to decrypt '%s'", from)
	}
	if err := s.Set(WithReason(ctx, fmt.Sprintf("Moved from %s to %s", from, to)), to, content); err != nil {
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
// recurse flag
func (s *Store) delete(ctx context.Context, name string, recurse bool) error {
	path := s.passfile(name)

	if recurse {
		if err := s.deleteRecurse(ctx, name, path); err != nil {
			return err
		}
	}
	if err := s.deleteSingle(ctx, path); err != nil {
		if !recurse {
			return err
		}
	}

	if err := s.rcs.Add(ctx, path); err != nil {
		if errors.Cause(err) == store.ErrGitNotInit {
			return nil
		}
		return errors.Wrapf(err, "failed to add '%s' to git", path)
	}

	if err := s.rcs.Commit(ctx, fmt.Sprintf("Remove %s from store.", name)); err != nil {
		if errors.Cause(err) == store.ErrGitNotInit {
			return nil
		}
		return errors.Wrapf(err, "failed to commit changes to git")
	}

	if !IsAutoSync(ctx) {
		return nil
	}

	if err := s.rcs.Push(ctx, "", ""); err != nil {
		if errors.Cause(err) == store.ErrGitNotInit || errors.Cause(err) == store.ErrGitNoRemote {
			return nil
		}
		return errors.Wrapf(err, "failed to push change to git remote")
	}

	return nil
}

func (s *Store) deleteRecurse(ctx context.Context, name, path string) error {
	if !s.storage.IsDir(ctx, name) && !s.storage.Exists(ctx, path) {
		return store.ErrNotFound
	}

	return s.storage.Prune(ctx, name)
}

func (s *Store) deleteSingle(ctx context.Context, path string) error {
	if !s.storage.Exists(ctx, path) {
		return store.ErrNotFound
	}

	return s.storage.Delete(ctx, path)
}
