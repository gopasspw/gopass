package leaf

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
)

// Copy will copy one entry to another location. Multi-store copies are
// supported. Each entry has to be decoded and encoded for the destination
// to make sure it's encrypted for the right set of recipients.
func (s *Store) Copy(ctx context.Context, from, to string) error {
	// recursive copy?
	if s.IsDir(ctx, from) {
		return fmt.Errorf("recursive operations are not supported")
	}

	content, err := s.Get(ctx, from)
	if err != nil {
		return fmt.Errorf("failed to get %q from store: %w", from, err)
	}

	if err := s.Set(ctxutil.WithCommitMessage(ctx, fmt.Sprintf("Copied from %s to %s", from, to)), to, content); err != nil {
		return fmt.Errorf("failed to save %q to store: %w", to, err)
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
		return fmt.Errorf("recursive operations are not supported")
	}

	content, err := s.Get(ctx, from)
	if err != nil {
		return fmt.Errorf("failed to decrypt %q: %w", from, err)
	}

	if err := s.Set(ctxutil.WithCommitMessage(ctx, fmt.Sprintf("Move from %s to %s", from, to)), to, content); err != nil {
		return fmt.Errorf("failed to write %q: %w", to, err)
	}

	if err := s.Delete(ctx, from); err != nil {
		return fmt.Errorf("failed to delete %q: %w", from, err)
	}

	return nil
}

// Delete will remove an single entry from the store.
func (s *Store) Delete(ctx context.Context, name string) error {
	return s.delete(ctx, name, false)
}

// Prune will remove a subtree from the Store.
func (s *Store) Prune(ctx context.Context, tree string) error {
	return s.delete(ctx, tree, true)
}

// delete will either delete one file or an directory tree depending on the
// recurse flag.
func (s *Store) delete(ctx context.Context, name string, recurse bool) error {
	path := s.passfile(name)

	if recurse {
		if err := s.deleteRecurse(ctx, name, path); err != nil {
			return err
		}
	}
	if err := s.deleteSingle(ctx, path); err != nil {
		// might fail if we deleted the root of a tree which isn't a secret
		// itself
		if !recurse {
			return err
		}
	}

	if !ctxutil.IsGitCommit(ctx) {
		return nil
	}

	if err := s.storage.Commit(ctx, fmt.Sprintf("Remove %s from store.", name)); err != nil {
		switch {
		case errors.Is(err, store.ErrGitNotInit):
			debug.Log("skipping git commit - git not initialized")
		case errors.Is(err, store.ErrGitNothingToCommit):
			debug.Log("skipping git commit - nothing to commit")
		default:
			return fmt.Errorf("failed to commit changes to git: %w", err)
		}
	}

	if err := s.storage.Push(ctx, "", ""); err != nil {
		if errors.Is(err, store.ErrGitNotInit) || errors.Is(err, store.ErrGitNoRemote) {
			return nil
		}

		return fmt.Errorf("failed to push change to git remote: %w", err)
	}

	return nil
}

func (s *Store) deleteRecurse(ctx context.Context, name, path string) error {
	if !s.storage.IsDir(ctx, name) && !s.storage.Exists(ctx, path) {
		return store.ErrNotFound
	}

	name = strings.TrimPrefix(name, string(filepath.Separator))

	debug.Log("Pruning %s", name)
	if err := s.storage.Prune(ctx, name); err != nil {
		debug.Log("storage.Prune(%v) failed", name)

		return err
	}

	if err := s.storage.Add(ctx, name); err != nil {
		if errors.Is(err, store.ErrGitNotInit) {
			return nil
		}

		return fmt.Errorf("failed to add %q to git: %w", path, err)
	}
	debug.Log("pruned")

	return nil
}

func (s *Store) deleteSingle(ctx context.Context, path string) error {
	if !s.storage.Exists(ctx, path) {
		return store.ErrNotFound
	}

	debug.Log("Deleting %s", path)
	if err := s.storage.Delete(ctx, path); err != nil {
		return err
	}

	if err := s.storage.Add(ctx, path); err != nil {
		if errors.Is(err, store.ErrGitNotInit) {
			return nil
		}

		return fmt.Errorf("failed to add %q to git: %w", path, err)
	}

	return nil
}
