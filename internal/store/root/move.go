package root

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/internal/store/leaf"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/pkg/errors"
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

func (r *Store) move(ctx context.Context, from, to string, delete bool) error {
	ctxFrom, subFrom, fromPrefix := r.getStore(ctx, from)
	ctxTo, subTo, _ := r.getStore(ctx, to)

	srcIsDir := r.IsDir(ctx, from)
	dstIsDir := r.IsDir(ctx, to)
	if srcIsDir && r.Exists(ctx, to) && !dstIsDir {
		return errors.New("destination is a file")
	}

	if err := r.moveFromTo(ctxFrom, ctxTo, subFrom, from, to, fromPrefix, srcIsDir, delete); err != nil {
		return err
	}
	if err := subFrom.Storage().Commit(ctxFrom, fmt.Sprintf("Move from %s to %s", from, to)); delete && err != nil {
		switch errors.Cause(err) {
		case store.ErrGitNotInit:
			debug.Log("skipping git commit - git not initialized")
		default:
			return errors.Wrapf(err, "failed to commit changes to git (from)")
		}
	}
	if !subFrom.Equals(subTo) {
		if err := subTo.Storage().Commit(ctxTo, fmt.Sprintf("Move from %s to %s", from, to)); err != nil {
			switch errors.Cause(err) {
			case store.ErrGitNotInit:
				debug.Log("skipping git commit - git not initialized")
			default:
				return errors.Wrapf(err, "failed to commit changes to git (to)")
			}
		}
	}

	if err := subFrom.Storage().Push(ctx, "", ""); err != nil {
		if errors.Cause(err) == store.ErrGitNotInit {
			msg := "Warning: git is not initialized for this.storage. Ignoring auto-push option\n" +
				"Run: gopass git init"
			out.Error(ctx, msg)
			return nil
		}
		if errors.Cause(err) == store.ErrGitNoRemote {
			msg := "Warning: git has no remote. Ignoring auto-push option\n" +
				"Run: gopass git remote add origin ..."
			out.Yellow(ctx, msg)
			return nil
		}
		return errors.Wrapf(err, "failed to push change to git remote")
	}
	if !subFrom.Equals(subTo) {
		if err := subTo.Storage().Push(ctx, "", ""); err != nil {
			if errors.Cause(err) == store.ErrGitNotInit {
				msg := "Warning: git is not initialized for this.storage. Ignoring auto-push option\n" +
					"Run: gopass git init"
				out.Error(ctx, msg)
				return nil
			}
			if errors.Cause(err) == store.ErrGitNoRemote {
				msg := "Warning: git has no remote. Ignoring auto-push option\n" +
					"Run: gopass git remote add origin ..."
				out.Yellow(ctx, msg)
				return nil
			}
			return errors.Wrapf(err, "failed to push change to git remote")
		}
	}
	return nil
}

func (r *Store) moveFromTo(ctxFrom, ctxTo context.Context, subFrom *leaf.Store, from, to, fromPrefix string, srcIsDir, delete bool) error {
	ctxFrom = ctxutil.WithGitCommit(ctxFrom, false)
	ctxTo = ctxutil.WithGitCommit(ctxTo, false)

	entries := []string{from}
	if r.IsDir(ctxFrom, from) {
		var err error
		entries, err = subFrom.List(ctxFrom, fromPrefix)
		if err != nil {
			return err
		}
	}
	if len(entries) < 1 {
		return errors.Errorf("no entries")
	}

	for _, src := range entries {
		dst := to
		if srcIsDir {
			// Follow the rsync convention to not re-create the source folder at the destination when a "/" is found
			if strings.HasSuffix(from, "/") {
				dst = path.Join(to, strings.TrimPrefix(src, from))
			} else {
				dst = path.Join(to, path.Base(from), strings.TrimPrefix(src, from))
			}
		}
		debug.Log("Moving %s (%s) => %s (%s) (sid:%t, delete:%t)\n", from, src, to, dst, srcIsDir, delete)

		content, err := r.Get(ctxFrom, src)
		if err != nil {
			return errors.Errorf("Source %s does not exist in source store %s: %s", from, subFrom.Alias(), err)
		}

		if err := r.Set(ctxutil.WithCommitMessage(ctxTo, fmt.Sprintf("Move from %s to %s", src, dst)), dst, content); err != nil {
			return errors.Wrapf(err, "failed to save secret '%s'", to)
		}

		if delete {
			debug.Log("Deleting %s from source %s", from, src)
			if err := r.Delete(ctxFrom, src); err != nil {
				return errors.Wrapf(err, "failed to delete secret '%s'", src)
			}
		}
	}
	return nil
}

// Delete will remove an single entry from the store
func (r *Store) Delete(ctx context.Context, name string) error {
	ctx, store, sn := r.getStore(ctx, name)
	if sn == "" {
		return errors.Errorf("can not delete a mount point. Use `gopass mounts remove %s`", store.Alias())
	}
	return store.Delete(ctx, sn)
}

// Prune will remove a subtree from the Store
func (r *Store) Prune(ctx context.Context, tree string) error {
	for mp := range r.mounts {
		if strings.HasPrefix(mp, tree) {
			return errors.Errorf("can not prune subtree with mounts. Unmount first: `gopass mounts remove %s`", mp)
		}
	}

	ctx, store, tree := r.getStore(ctx, tree)
	return store.Prune(ctx, tree)
}
