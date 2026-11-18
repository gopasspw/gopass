package root

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
)

// Fsck checks the entries under path in the store it belongs to, or every
// store when path is empty. A non-empty store limits fsck to that mount point.
func (r *Store) Fsck(ctx context.Context, store, path string, progress ctxutil.ProgressCallback) error {
	var result []error

	mp := r.MountPoint(path)
	subPath := strings.TrimPrefix(strings.TrimPrefix(path, mp), "/")

	if store != "" {
		if _, found := r.mounts[store]; !found {
			return fmt.Errorf("no such mount point %q", store)
		}

		if path != "" && mp != store {
			return fmt.Errorf("%q is not in mount point %q", path, store)
		}
	}

	for alias, sub := range r.mounts {
		if sub == nil {
			continue
		}

		if store != "" && alias != store {
			continue
		}

		if path != "" && alias != mp {
			continue
		}

		// check sub store
		debug.Log("Checking mount point %s", alias)

		if err := sub.Fsck(ctx, subPath, progress); err != nil {
			out.Errorf(ctx, "Fsck failed on sub store %s: %s", alias, err)
			result = append(result, err)
		}

		debug.Log("Checked mount point %s", alias)
	}

	if store != "" || mp != "" {
		return errors.Join(result...)
	}

	// check root store
	debug.Log("Checking root store")
	if err := r.store.Fsck(ctx, path, progress); err != nil {
		out.Errorf(ctx, "Fsck failed on root store: %s", err)
		result = append(result, err)
	}

	debug.Log("Checked root store")

	return errors.Join(result...)
}
