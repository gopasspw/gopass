package root

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/internal/store/leaf"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
)

// Copy will copy one entry to another location. Multi-store copies are
// supported. Each entry has to be decoded and encoded for the destination
// to make sure it's encrypted for the right set of recipients.
func (r *Store) Copy(ctx context.Context, from, to string) error {
	return r.move(ctx, from, to, false)
}

// Move will move one entry from one location to another. Cross-store moves are
// supported. Moving an entry will decode it from the old location, encode it
// for the destination store with the right set of recipients and remove it
// from the old location afterwards.
func (r *Store) Move(ctx context.Context, from, to string) error {
	return r.move(ctx, from, to, true)
}

func (r *Store) move(ctx context.Context, from, to string, del bool) error {
	subFrom, fromPrefix := r.getStore(from)
	subTo, _ := r.getStore(to)

	srcIsDir := r.IsDir(ctx, from)
	dstIsDir := r.IsDir(ctx, to)
	if srcIsDir && r.Exists(ctx, to) && !dstIsDir {
		return fmt.Errorf("destination is a file")
	}

	if err := r.moveFromTo(ctx, subFrom, from, to, fromPrefix, srcIsDir, dstIsDir, del); err != nil {
		return err
	}

	if err := subFrom.Storage().Commit(ctx, fmt.Sprintf("Move from %s to %s", from, to)); del && err != nil {
		switch {
		case errors.Is(err, store.ErrGitNotInit):
			debug.Log("skipping git commit - git not initialized")
		default:
			return fmt.Errorf("failed to commit changes to git (from): %w", err)
		}
	}

	if !subFrom.Equals(subTo) {
		if err := subTo.Storage().Commit(ctx, fmt.Sprintf("Move from %s to %s", from, to)); err != nil {
			switch errors.Unwrap(err) {
			case store.ErrGitNotInit:
				debug.Log("skipping git commit - git not initialized")
			default:
				return fmt.Errorf("failed to commit changes to git (to): %w", err)
			}
		}
	}

	if err := subFrom.Storage().Push(ctx, "", ""); err != nil {
		if errors.Is(err, store.ErrGitNotInit) {
			msg := "Warning: git is not initialized for this storage. Ignoring auto-push option\n" +
				"Run: gopass git init"
			debug.Log(msg)
			return nil
		}
		if errors.Is(err, store.ErrGitNoRemote) {
			msg := "Warning: git has no remote. Ignoring auto-push option\n" +
				"Run: gopass git remote add origin ..."
			debug.Log(msg)
			return nil
		}
		return fmt.Errorf("failed to push change to git remote: %w", err)
	}

	if !subFrom.Equals(subTo) {
		if err := subTo.Storage().Push(ctx, "", ""); err != nil {
			if errors.Is(err, store.ErrGitNotInit) {
				msg := "Warning: git is not initialized for this storage. Ignoring auto-push option\n" +
					"Run: gopass git init"
				debug.Log(msg)
				return nil
			}
			if errors.Is(err, store.ErrGitNoRemote) {
				msg := "Warning: git has no remote. Ignoring auto-push option\n" +
					"Run: gopass git remote add origin ..."
				debug.Log(msg)
				return nil
			}
			return fmt.Errorf("failed to push change to git remote: %w", err)
		}
	}

	return nil
}

func (r *Store) moveFromTo(ctx context.Context, subFrom *leaf.Store, from, to, fromPrefix string, srcIsDir, dstIsDir, del bool) error {
	ctx = ctxutil.WithGitCommit(ctx, false)

	entries := []string{from}
	// if the source is a directory we enumerate all it's children
	// and move them one by one.
	if r.IsDir(ctx, from) {
		var err error
		entries, err = subFrom.List(ctx, fromPrefix)
		if err != nil {
			return err
		}
	}
	if len(entries) < 1 {
		return fmt.Errorf("no entries")
	}

	debug.Log("Moving %q to %q (entries: %+v)", from, to, entries)

	for _, src := range entries {
		dst := to
		if srcIsDir {
			// Follow the rsync convention to not re-create the source folder at the destination when a "/" is found
			if strings.HasSuffix(from, "/") {
				dst = path.Join(to, strings.TrimPrefix(src, from))
			} else {
				dst = path.Join(to, path.Base(from), strings.TrimPrefix(src, from))
			}
		} else {
			if dstIsDir || strings.HasSuffix(to, "/") {
				dst = path.Join(to, path.Base(src))
			}
		}
		debug.Log("Moving %q (%q) => %q (%q) (sid:%t, did:%t, delete:%t)\n", from, src, to, dst, srcIsDir, dstIsDir, del)

		content, err := r.Get(ctx, src)
		if err != nil {
			return fmt.Errorf("source %s does not exist in source store %s: %s", from, subFrom.Alias(), err)
		}

		if err := r.Set(ctxutil.WithCommitMessage(ctx, fmt.Sprintf("Move from %s to %s", src, dst)), dst, content); err != nil {
			return fmt.Errorf("failed to save secret %q: %w", to, err)
		}

		if del {
			debug.Log("Deleting %s from source %s", from, src)
			if err := r.Delete(ctx, src); err != nil {
				return fmt.Errorf("failed to delete secret %q: %w", src, err)
			}
		}
	}
	return nil
}

// Delete will remove an single entry from the store.
func (r *Store) Delete(ctx context.Context, name string) error {
	store, sn := r.getStore(name)
	if sn == "" {
		return fmt.Errorf("can not delete a mount point. Use `gopass mounts remove %s`", store.Alias())
	}
	return store.Delete(ctx, sn)
}

// Prune will remove a subtree from the Store.
func (r *Store) Prune(ctx context.Context, tree string) error {
	for mp := range r.mounts {
		if strings.HasPrefix(mp, tree) {
			return fmt.Errorf("can not prune subtree with mounts. Unmount first: `gopass mounts remove %s`", mp)
		}
	}

	store, tree := r.getStore(tree)
	return store.Prune(ctx, tree)
}
