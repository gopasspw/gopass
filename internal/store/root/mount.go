package root

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/internal/store/leaf"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/fsutil"
)

// AddMount adds a new mount.
func (r *Store) AddMount(ctx context.Context, alias, path string, keys ...string) error {
	if err := r.addMount(ctx, alias, path, keys...); err != nil {
		return fmt.Errorf("failed to add mount: %w", err)
	}

	// check for duplicate mounts
	return r.checkMounts()
}

func (r *Store) addMount(ctx context.Context, alias, path string, keys ...string) error {
	if alias == "" {
		return fmt.Errorf("alias must not be empty")
	}

	if r.mounts == nil {
		r.mounts = make(map[string]*leaf.Store, 1)
	}

	if _, found := r.mounts[alias]; found {
		return AlreadyMountedError(alias)
	}

	fullPath := fsutil.CleanPath(path)
	debug.Log("addMount - Path: %s - Full: %s", path, fullPath)

	// initialize sub store
	s, err := r.initSub(ctx, alias, fullPath, keys)
	if err != nil {
		return fmt.Errorf("failed to init sub store %q at %q: %w", alias, fullPath, err)
	}

	r.mounts[alias] = s
	if err := r.cfg.SetMountPath(alias, path); err != nil {
		return fmt.Errorf("failed to set mount path: %w", err)
	}

	debug.Log("Added mount %s -> %s (%s)", alias, path, fullPath)

	return nil
}

func (r *Store) initSub(ctx context.Context, alias, path string, keys []string) (*leaf.Store, error) {
	// init regular sub store
	s, err := leaf.New(ctx, alias, path)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize store %q at %q: %w", alias, path, err)
	}

	if s.IsInitialized(ctx) {
		return s, nil
	}

	debug.Log("[%s] Mount %s is not initialized", alias, path)

	if len(keys) < 1 {
		debug.Log("[%s] No keys available", alias)

		return s, NotInitializedError{alias, path}
	}

	debug.Log("[%s] Trying to initialize at %s for %+v", alias, path, keys)

	if err := s.Init(ctx, path, keys...); err != nil {
		return s, fmt.Errorf("failed to initialize store %q at %q: %w", alias, path, err)
	}

	out.Printf(ctx, "Password store %s initialized for:", path)

	for _, r := range s.Recipients(ctx) {
		out.Noticef(ctx, "  %s", r)
	}

	return s, nil
}

// RemoveMount removes and existing mount.
func (r *Store) RemoveMount(ctx context.Context, alias string) error {
	if _, found := r.mounts[alias]; !found {
		out.Warningf(ctx, "%s is not mounted", alias)
	}

	if _, found := r.mounts[alias]; !found {
		out.Warningf(ctx, "%s is not initialized", alias)
	}

	delete(r.mounts, alias)
	if err := r.cfg.Unset("", "mounts."+alias+".path"); err != nil {
		return err
	}

	return nil
}

// Mounts returns a map of mounts with their paths.
func (r *Store) Mounts() map[string]string {
	m := make(map[string]string, len(r.mounts))
	for alias, sub := range r.mounts {
		m[alias] = sub.Path()
	}

	return m
}

// MountPoints returns a sorted list of mount points. It encodes the logic that
// the longer a mount point the more specific it is. This allows to "shadow" a
// shorter mount point by a longer one.
func (r *Store) MountPoints() []string {
	mps := make([]string, 0, len(r.mounts))
	for k := range r.mounts {
		mps = append(mps, k)
	}

	sort.Sort(sort.Reverse(store.ByPathLen(mps)))

	return mps
}

// MountPoint returns the most-specific mount point for the given key.
func (r *Store) MountPoint(name string) string {
	for _, mp := range r.MountPoints() {
		if strings.HasPrefix(name+"/", mp+"/") {
			return mp
		}
	}

	return ""
}

// Lock drops all cached credentials, if any. Mostly only useful
// for the gopass REPL.
func (r *Store) Lock() error {
	for _, sub := range r.mounts {
		if err := sub.Lock(); err != nil {
			return err
		}
	}

	return r.store.Lock()
}

// getStore returns the Store object at the most-specific mount point for the
// given key. returns sub store reference, truncated path to secret.
func (r *Store) getStore(name string) (*leaf.Store, string) {
	name = strings.TrimSuffix(name, "/")
	mp := r.MountPoint(name)

	if sub, found := r.mounts[mp]; found {
		return sub, strings.TrimPrefix(name, sub.Alias())
	}

	return r.store, name
}

// GetSubStore returns an exact match for a mount point or an error if this
// mount point does not exist.
func (r *Store) GetSubStore(name string) (*leaf.Store, error) {
	if name == "" {
		return r.store, nil
	}

	if sub, found := r.mounts[name]; found {
		return sub, nil
	}

	debug.Log("mounts available: %+v", r.mounts)

	return nil, fmt.Errorf("no such mount point %q", name)
}

// checkMounts performs some sanity checks on our mounts. At the moment it
// only checks if some path is mounted twice.
func (r *Store) checkMounts() error {
	paths := make(map[string]string, len(r.mounts))
	for k, v := range r.mounts {
		if _, found := paths[v.Path()]; found {
			return fmt.Errorf("doubly mounted path at %s: %s", v.Path(), k)
		}

		paths[v.Path()] = k
	}

	return nil
}
