package root

import (
	"context"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/config"
	"github.com/justwatchcom/gopass/store"
	"github.com/justwatchcom/gopass/store/sub"
	"github.com/justwatchcom/gopass/utils/fsutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/pkg/errors"
)

// AddMount adds a new mount
func (r *Store) AddMount(ctx context.Context, alias, path string, keys ...string) error {
	path = fsutil.CleanPath(path)
	if err := r.addMount(ctx, alias, path, keys...); err != nil {
		return errors.Wrapf(err, "failed to add mount")
	}

	// check for duplicate mounts
	return r.checkMounts()
}

func (r *Store) addMount(ctx context.Context, alias, path string, keys ...string) error {
	if alias == "" {
		return errors.Errorf("alias must not be empty")
	}
	if r.mounts == nil {
		r.mounts = make(map[string]*sub.Store, 1)
	}
	if _, found := r.mounts[alias]; found {
		return errors.Errorf("%s is already mounted", alias)
	}

	// propagate our config settings to the sub store
	s := sub.New(alias, path, r.gpg)

	if !s.Initialized() {
		if len(keys) < 1 {
			return errors.Errorf("password store %s is not initialized. Try gopass init --store %s --path %s", alias, alias, path)
		}
		if err := s.Init(ctx, path, keys...); err != nil {
			return errors.Wrapf(err, "failed to initialize store '%s' at '%s'", alias, path)
		}
		out.Green(ctx, "Password store %s initialized for:", path)
		for _, r := range s.Recipients(ctx) {
			color.Yellow(r)
		}
	}

	r.mounts[alias] = s
	if r.cfg.Mounts == nil {
		r.cfg.Mounts = make(map[string]*config.StoreConfig, 1)
	}
	// imporant: copy root config to avoid overwriting it with sub store
	// values
	sc := *r.cfg.Root
	sc.Path = path
	r.cfg.Mounts[alias] = &sc
	return nil
}

// RemoveMount removes and existing mount
func (r *Store) RemoveMount(ctx context.Context, alias string) error {
	if _, found := r.mounts[alias]; !found {
		return errors.Errorf("%s is not mounted", alias)
	}
	if _, found := r.mounts[alias]; !found {
		out.Yellow(ctx, "%s is not initialized", alias)
	}
	delete(r.mounts, alias)
	delete(r.cfg.Mounts, alias)
	return nil
}

// Mounts returns a map of mounts with their paths
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

// mountPoint returns the most-specific mount point for the given key
func (r *Store) mountPoint(name string) string {
	for _, mp := range r.MountPoints() {
		if strings.HasPrefix(name+"/", mp+"/") {
			return mp
		}
	}
	return ""
}

// getStore returns the Store object at the most-specific mount point for the
// given key
// context with sub store options set, sub store reference, truncated path to secret
func (r *Store) getStore(ctx context.Context, name string) (context.Context, *sub.Store, string) {
	name = strings.TrimSuffix(name, "/")
	mp := r.mountPoint(name)
	if sub, found := r.mounts[mp]; found {
		return r.cfg.Mounts[mp].WithContext(ctx), sub, strings.TrimPrefix(name, sub.Alias())
	}
	return ctx, r.store, name
}

// GetSubStore returns an exact match for a mount point or an error if this
// mount point does not exist
func (r *Store) GetSubStore(name string) (*sub.Store, error) {
	if name == "" {
		return r.store, nil
	}
	if sub, found := r.mounts[name]; found {
		return sub, nil
	}
	return nil, errors.Errorf("no such mount point '%s'", name)
}

// checkMounts performs some sanity checks on our mounts. At the moment it
// only checks if some path is mounted twice.
func (r *Store) checkMounts() error {
	paths := make(map[string]string, len(r.mounts))
	for k, v := range r.mounts {
		if _, found := paths[v.Path()]; found {
			return errors.Errorf("Doubly mounted path at %s: %s", v.Path(), k)
		}
		paths[v.Path()] = k
	}
	return nil
}
