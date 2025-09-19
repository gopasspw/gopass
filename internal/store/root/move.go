package root

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/internal/store/leaf"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/fsutil"
)

// Copy will copy one entry to another location. Multi-store copies are
// supported. Each entry has to be decoded and encoded for the destination
// to make sure it's encrypted for the right set of recipients.
func (r *Store) Copy(ctx context.Context, from, to string) error {
	debug.Log("Copy %s to %s", from, to)

	return r.move(ctx, from, to, false)
}

// Move will move one entry from one location to another. Cross-store moves are
// supported. Moving an entry will decode it from the old location, encode it
// for the destination store with the right set of recipients and remove it
// from the old location afterwards.
func (r *Store) Move(ctx context.Context, from, to string) error {
	debug.Log("Move %s to %s", from, to)

	return r.move(ctx, from, to, true)
}

// move handles both copy and move operations. Since the only difference is
// deleting the source entry after the copy, we can reuse the same code.
func (r *Store) move(ctx context.Context, from, to string, del bool) error {
	subFrom, fromPrefix := r.getStore(from)
	subTo, _ := r.getStore(to)

	if err := r.moveFromTo(ctx, subFrom, from, to, fromPrefix, del); err != nil {
		return err
	}

	commitMsg := ctxutil.GetCommitMessage(ctx)
	if err := subFrom.Storage().TryCommit(ctx, commitMsg); del && err != nil {
		return fmt.Errorf("failed to commit changes to git (%s): %w", subFrom.Alias(), err)
	}

	if !subFrom.Equals(subTo) {
		if err := subTo.Storage().TryCommit(ctx, commitMsg); err != nil {
			return fmt.Errorf("failed to commit changes to git (%s): %w", subTo.Alias(), err)
		}
	}

	if err := subFrom.Storage().TryPush(ctx, "", ""); err != nil {
		return fmt.Errorf("failed to push change to git remote: %w", err)
	}

	if subFrom.Equals(subTo) {
		return nil
	}

	if err := subTo.Storage().TryPush(ctx, "", ""); err != nil {
		return fmt.Errorf("failed to push change to git remote: %w", err)
	}

	return nil
}

func (r *Store) moveFromTo(ctx context.Context, subFrom *leaf.Store, from, to, fromPrefix string, del bool) error {
	ctx = ctxutil.WithGitCommit(ctx, false)

	// source is a directory and not a "shadowed" leaf
	srcIsDir := r.IsDir(ctx, from) && !r.Exists(ctx, from)
	dstIsDir := r.IsDir(ctx, to)

	if srcIsDir && r.Exists(ctx, to) && !dstIsDir {
		return fmt.Errorf("destination is a file")
	}

	entries := []string{from}
	// if the source is a directory we enumerate all it's children
	// and move them one by one.
	if srcIsDir {
		var err error

		entries, err = subFrom.List(ctx, fromPrefix+"/")
		if err != nil {
			return err
		}
	}

	if len(entries) < 1 {
		debug.Log("Subtree %q has no entries", from)

		return fmt.Errorf("no entries")
	}

	debug.Log("Moving (sub) tree %q to %q (entries: %+v)", from, to, entries)

	var moved uint
	for _, src := range entries {
		dst := computeMoveDestination(src, from, to, srcIsDir, dstIsDir)
		if src == dst {
			debug.Log("skipping %q. src eq dst", src)

			continue
		}
		debug.Log("Moving entry %q (%q) => %q (%q) (srcIsDir:%t, dstIsDir:%t, delete:%t)\n", src, from, dst, to, srcIsDir, dstIsDir, del)

		err := r.directMove(ctx, src, dst, del)
		if err == nil {
			moved++
			debug.Log("directly moved from %q to %q", src, dst)

			continue
		}

		debug.Log("direct move failed to move entry %q to %q: %s. Falling back to get and set", src, dst, err)

		content, err := r.Get(ctx, src)
		if err != nil {
			return fmt.Errorf("source %s does not exist in source store %s: %w", from, subFrom.Alias(), err)
		}

		if err := r.Set(ctx, dst, content); err != nil {
			if !errors.Is(err, store.ErrMeaninglessWrite) {
				return fmt.Errorf("failed to save secret %q to store: %w", to, err)
			}
			out.Warningf(ctx, "No need to write: the secret is already there and with the right value")
		}

		if del {
			debug.Log("Deleting moved entry %q from source %q", from, src)

			if err := r.Delete(ctx, src); err != nil {
				return fmt.Errorf("failed to delete secret %q: %w", src, err)
			}
		}

		moved++
	}

	if moved < 1 {
		return fmt.Errorf("no entries moved")
	}

	debug.Log("Moved (sub) tree %q to %q", from, to)

	return nil
}

func (r *Store) directMove(ctx context.Context, from, to string, del bool) error {
	debug.Log("directMove from %q to %q", from, to)

	// will also remove the store prefix, if applicable
	subFrom, from := r.getStore(from)

	// we don't remove store prefix for destination, as it can be a new folder
	subTo, to := r.getStore(to)

	if subFrom.Equals(subTo) {
		debug.Log("directMove from %q to %q: same store", from, to)

		if del {
			return subFrom.Move(ctx, from, to)
		}

		return subFrom.Copy(ctx, from, to)
	}

	debug.Log("cross mount direct move from %s%s to %s%s", subFrom.Alias(), from, subTo.Alias(), to)

	// assemble source and destination paths, call fsutil.CopyFile(from, to), remove source
	// if del is true and then git add and commit both stores.
	sfn := filepath.Join(subFrom.Path(), subFrom.Passfile(from))
	dfn := filepath.Join(subTo.Path(), subTo.Passfile(to))

	if err := fsutil.CopyFile(sfn, dfn); err != nil {
		return fmt.Errorf("failed to copy %q to %q: %w", from, to, err)
	}

	if del {
		if err := os.Remove(sfn); err != nil {
			return fmt.Errorf("failed to delete %q from %s: %w", sfn, subFrom.Alias(), err)
		}
	}

	if err := subFrom.Storage().Add(ctx, sfn); err != nil {
		debug.Log("failed to add %q to %s: %w", sfn, subFrom.Alias(), err)
	}

	if err := subTo.Storage().Add(ctx, dfn); err != nil {
		debug.Log("failed to add %q to %s: %w", dfn, subTo.Alias(), err)
	}

	return nil
}

func computeMoveDestination(src, from, to string, srcIsDir, dstIsDir bool) string {
	// special case: moving up to the root
	if to == "." || to == "/" {
		dstIsDir = false
		to = ""
	}

	// are we moving into an existing directory? Then we just need to prepend
	// it's name to the source.
	// a -> b
	// - a/f1 -> b/a/f1
	// a -> b
	// - a -> b/a
	if dstIsDir {
		if !srcIsDir {
			return path.Join(to, path.Base(src))
		}

		return path.Join(to, src)
	}

	// are we moving a simple file? that's easy
	if !srcIsDir {
		// otherwise we just rename a file to another name
		return to
	}

	// move a/ b, where a is a directory with a trailing slash and b
	// does not exist, i.e. move a to b
	if strings.HasSuffix(from, "/") {
		return path.Join(to, strings.TrimPrefix(src, from))
	}

	// move a b, where a is a directory but not b, i.e. rename a to b.
	// this is applied to every child of a, so we need to remove the
	// old prefix (a) and add the new one (b).
	return path.Join(to, strings.TrimPrefix(src, from))
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
