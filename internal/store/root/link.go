package root

import (
	"context"
	"fmt"
)

// Link creates a symlink
func (r *Store) Link(ctx context.Context, from, to string) error {
	subFrom, fName := r.getStore(from)
	subTo, tName := r.getStore(to)

	if !subFrom.Equals(subTo) {
		return fmt.Errorf("sylinks across stores are not supported")
	}

	return subFrom.Link(ctx, fName, tName)
}
