package root

import (
	"context"
	"fmt"

	"github.com/gopasspw/gopass/pkg/gopass"
)

// Set encodes and write the ciphertext of one entry to disk.
func (r *Store) Set(ctx context.Context, name string, sec gopass.Byter) error {
	store, name := r.getStore(name)

	return store.Set(ctx, name, sec)
}

// SameRecipientScope reports whether writes to two names use the same store and recipient file.
func (r *Store) SameRecipientScope(ctx context.Context, a, b string) bool {
	subA, a := r.getStore(a)
	subB, b := r.getStore(b)

	return subA != nil && subA == subB && subA.SameRecipientScope(ctx, a, b)
}

// CheckImportWrite checks whether a batch can replace a secret without
// overwriting staged user changes.
func (r *Store) CheckImportWrite(ctx context.Context, name string) error {
	sub, path := r.getStore(name)
	if sub == nil || path == "" {
		return fmt.Errorf("no store for import target %q", name)
	}

	return sub.CheckImportWrite(ctx, path)
}

// StageImport stages the import's fresh and retried files in one mount.
func (r *Store) StageImport(ctx context.Context, mount string, names, updated []string) (bool, error) {
	sub, err := r.GetSubStore(mount)
	if err != nil {
		return false, err
	}
	relative := func(names []string) ([]string, error) {
		paths := make([]string, 0, len(names))
		for _, name := range names {
			target, path := r.getStore(name)
			if sub == nil || sub != target || path == "" {
				return nil, fmt.Errorf("secret %q is not in mount %q", name, mount)
			}
			paths = append(paths, path)
		}

		return paths, nil
	}
	paths, err := relative(names)
	if err != nil {
		return false, err
	}
	updatedPaths, err := relative(updated)
	if err != nil {
		return false, err
	}

	return sub.StageImport(ctx, paths, updatedPaths)
}
