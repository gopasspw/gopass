package root

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store/leaf"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
)

// Fsck checks all stores/entries matching the given store and prefix.
func (r *Store) Fsck(ctx context.Context, store, path string, progress ctxutil.ProgressCallback) error {
	if store != "" {
		return r.fsckStore(ctx, store, path, progress)
	}

	var result []error

	cleanPath := strings.Trim(path, "/")
	mp := ""
	if cleanPath != "" {
		mp = r.MountPoint(cleanPath)
	}

	matchedMount := false

	for _, alias := range r.MountPoints() {
		sub := r.mounts[alias]
		if sub == nil {
			continue
		}

		subPath, ok := matchMountFsckPath(alias, mp, cleanPath)
		if !ok {
			continue
		}

		matchedMount = true

		debug.Log("Checking mount point %s", alias)

		if err := sub.Fsck(ctx, subPath, progress); err != nil {
			out.Errorf(ctx, "Fsck failed on sub store %s: %s", alias, err)
			result = append(result, err)
		}

		debug.Log("Checked mount point %s", alias)
	}

	if r.shouldFsckRoot(ctx, mp, cleanPath, matchedMount) {
		debug.Log("Checking root store")

		if err := r.store.Fsck(ctx, cleanPath, progress); err != nil {
			out.Errorf(ctx, "Fsck failed on root store: %s", err)
			result = append(result, err)
		}

		debug.Log("Checked root store")
	}

	return errors.Join(result...)
}

func (r *Store) fsckStore(ctx context.Context, store, path string, progress ctxutil.ProgressCallback) error {
	cleanPath := strings.Trim(path, "/")

	if store == "<root>" {
		if r.store == nil {
			return fmt.Errorf("root store is not initialized")
		}

		debug.Log("Checking root store")

		if err := r.store.Fsck(ctx, cleanPath, progress); err != nil {
			out.Errorf(ctx, "Fsck failed on root store: %s", err)

			return err
		}

		debug.Log("Checked root store")

		return nil
	}

	alias := CleanMountAlias(store)
	if alias == "" {
		return fmt.Errorf("no such mount point %q", store)
	}

	sub, err := r.GetSubStore(alias)
	if err != nil {
		return err
	}

	subPath := resolveSubStorePath(ctx, sub, alias, cleanPath)

	debug.Log("Checking mount point %s", alias)

	if err := sub.Fsck(ctx, subPath, progress); err != nil {
		out.Errorf(ctx, "Fsck failed on sub store %s: %s", alias, err)

		return err
	}

	debug.Log("Checked mount point %s", alias)

	return nil
}

func resolveSubStorePath(ctx context.Context, sub *leaf.Store, alias, cleanPath string) string {
	if cleanPath == "" || cleanPath == alias {
		return ""
	}

	if sub.Exists(ctx, cleanPath) || sub.IsDir(ctx, cleanPath) {
		return cleanPath
	}

	if after, ok := strings.CutPrefix(cleanPath, alias+"/"); ok {
		return after
	}

	return cleanPath
}

func matchMountFsckPath(alias, mp, cleanPath string) (string, bool) {
	if cleanPath == "" {
		return "", true
	}

	if alias == mp {
		return strings.TrimPrefix(strings.TrimPrefix(cleanPath, alias), "/"), true
	}

	if strings.HasPrefix(alias, cleanPath+"/") {
		return "", true
	}

	return "", false
}

func (r *Store) shouldFsckRoot(ctx context.Context, mp, cleanPath string, matchedMount bool) bool {
	if r.store == nil {
		return false
	}

	if cleanPath == "" {
		return true
	}

	if mp != "" {
		return false
	}

	if !matchedMount {
		return true
	}

	return r.store.Exists(ctx, cleanPath) || r.store.IsDir(ctx, cleanPath)
}
