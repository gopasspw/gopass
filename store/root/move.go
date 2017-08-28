package root

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// Copy will copy one entry to another location. Multi-store copies are
// supported. Each entry has to be decoded and encoded for the destination
// to make sure it's encrypted for the right set of recipients.
func (r *Store) Copy(from, to string) error {
	subFrom := r.getStore(from)
	subTo := r.getStore(to)

	from = strings.TrimPrefix(from, subFrom.Alias())
	to = strings.TrimPrefix(to, subFrom.Alias())

	// cross-store copy
	if !subFrom.Equals(subTo) {
		content, err := subFrom.Get(from)
		if err != nil {
			return errors.Wrapf(err, "failed to retrieve secret '%s'", from)
		}
		if err := subTo.Set(to, content, fmt.Sprintf("Copied from %s to %s", from, to)); err != nil {
			return errors.Wrapf(err, "failed to store secret '%s'", to)
		}
		return nil
	}

	return subFrom.Copy(from, to)
}

// Move will move one entry from one location to another. Cross-store moves are
// supported. Moving an entry will decode it from the old location, encode it
// for the destination store with the right set of recipients and remove it
// from the old location afterwards.
func (r *Store) Move(from, to string) error {
	subFrom := r.getStore(from)
	subTo := r.getStore(to)

	from = strings.TrimPrefix(from, subFrom.Alias())

	// cross-store move
	if !subFrom.Equals(subTo) {
		to = strings.TrimPrefix(to, subTo.Alias())
		content, err := subFrom.Get(from)
		if err != nil {
			return errors.Errorf("Source %s does not exist in source store %s: %s", from, subFrom.Alias(), err)
		}
		if err := subTo.Set(to, content, fmt.Sprintf("Moved from %s to %s", from, to)); err != nil {
			return errors.Wrapf(err, "failed to save secret '%s'", to)
		}
		if err := subFrom.Delete(from); err != nil {
			return errors.Wrapf(err, "failed to delete secret '%s'", from)
		}
		return nil
	}

	to = strings.TrimPrefix(to, subFrom.Alias())
	return subFrom.Move(from, to)
}

// Delete will remove an single entry from the store
func (r *Store) Delete(name string) error {
	store := r.getStore(name)
	sn := strings.TrimPrefix(name, store.Alias())
	if sn == "" {
		return errors.Errorf("can not delete a mount point. Use `gopass mount remove %s`", store.Alias())
	}
	return store.Delete(sn)
}

// Prune will remove a subtree from the Store
func (r *Store) Prune(tree string) error {
	for mp := range r.mounts {
		if strings.HasPrefix(mp, tree) {
			return errors.Errorf("can not prune subtree with mounts. Unmount first: `gopass mount remove %s`", mp)
		}
	}

	store := r.getStore(tree)
	return store.Prune(strings.TrimPrefix(tree, store.Alias()))
}
