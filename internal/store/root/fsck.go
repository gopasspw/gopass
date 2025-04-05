package root

import (
	"context"
	"errors"
	"strings"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/debug"
)

// Fsck checks all stores/entries matching the given prefix.
func (r *Store) Fsck(ctx context.Context, store, path string) error {
	var result []error

	for alias, sub := range r.mounts {
		if sub == nil {
			continue
		}

		if store != "" && alias != store {
			continue
		}

		if path != "" && !strings.HasPrefix(path, alias+"/") {
			continue
		}

		path = strings.TrimPrefix(path, alias+"/")

		// check sub store
		debug.Log("Checking mount point %s", alias)

		if err := sub.Fsck(ctx, path); err != nil {
			out.Errorf(ctx, "fsck failed on sub store %s: %s", alias, err)
			result = append(result, err)
		}

		debug.Log("Checked mount point %s", alias)
	}

	// check root store
	debug.Log("Checking root store")
	if err := r.store.Fsck(ctx, path); err != nil {
		out.Errorf(ctx, "fsck failed on root store: %s", err)
		result = append(result, err)
	}

	debug.Log("Checked root store")

	return errors.Join(result...)
}
