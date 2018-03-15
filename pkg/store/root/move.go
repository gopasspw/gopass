package root

import (
	"context"
	"fmt"
	"strings"

	"github.com/justwatchcom/gopass/pkg/store/sub"
	"github.com/pkg/errors"
)

// Copy will copy one entry to another location. Multi-store copies are
// supported. Each entry has to be decoded and encoded for the destination
// to make sure it's encrypted for the right set of recipients.
func (r *Store) Copy(ctx context.Context, from, to string) error {
	ctxFrom, subFrom, from := r.getStore(ctx, from)
	ctxTo, subTo, _ := r.getStore(ctx, to)

	to = strings.TrimPrefix(to, subFrom.Alias())

	// cross-store copy
	if !subFrom.Equals(subTo) {
		content, err := subFrom.Get(ctxFrom, from)
		if err != nil {
			return errors.Wrapf(err, "failed to retrieve secret '%s'", from)
		}
		if err := subTo.Set(sub.WithReason(ctxTo, fmt.Sprintf("Copied from %s to %s", from, to)), to, content); err != nil {
			return errors.Wrapf(err, "failed to store secret '%s'", to)
		}
		return nil
	}

	return subFrom.Copy(ctxFrom, from, to)
}

// Move will move one entry from one location to another. Cross-store moves are
// supported. Moving an entry will decode it from the old location, encode it
// for the destination store with the right set of recipients and remove it
// from the old location afterwards.
func (r *Store) Move(ctx context.Context, from, to string) error {
	ctxFrom, subFrom, from := r.getStore(ctx, from)
	ctxTo, subTo, _ := r.getStore(ctx, to)

	// cross-store move
	if !subFrom.Equals(subTo) {
		to = strings.TrimPrefix(to, subTo.Alias())
		content, err := subFrom.Get(ctxFrom, from)
		if err != nil {
			return errors.Errorf("Source %s does not exist in source store %s: %s", from, subFrom.Alias(), err)
		}
		if err := subTo.Set(sub.WithReason(ctxTo, fmt.Sprintf("Moved from %s to %s", from, to)), to, content); err != nil {
			return errors.Wrapf(err, "failed to save secret '%s'", to)
		}
		if err := subFrom.Delete(ctxFrom, from); err != nil {
			return errors.Wrapf(err, "failed to delete secret '%s'", from)
		}
		return nil
	}

	to = strings.TrimPrefix(to, subFrom.Alias())
	return subFrom.Move(ctxFrom, from, to)
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
