package root

import (
	"fmt"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/fsutil"
	"github.com/justwatchcom/gopass/store"
	"github.com/justwatchcom/gopass/store/sub"
	"github.com/pkg/errors"
)

// AddMount adds a new mount
func (r *Store) AddMount(alias, path string, keys ...string) error {
	path = fsutil.CleanPath(path)
	if _, found := r.mounts[alias]; found {
		return errors.Errorf("%s is already mounted", alias)
	}
	if err := r.addMount(alias, path, keys...); err != nil {
		return errors.Wrapf(err, "failed to add mount")
	}

	// check for duplicate mounts
	return r.checkMounts()
}

func (r *Store) addMount(alias, path string, keys ...string) error {
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
	cfg := r.Config()
	cfg.Path = fsutil.CleanPath(path)
	s, err := sub.New(alias, cfg)
	if err != nil {
		return errors.Wrapf(err, "failed to create new sub store '%s'", alias)
	}

	if !s.Initialized() {
		if len(keys) < 1 {
			return errors.Errorf("password store %s is not initialized. Try gopass init --store %s --path %s", alias, alias, path)
		}
		if err := s.Init(path, keys...); err != nil {
			return errors.Wrapf(err, "failed to initialize store '%s' at '%s'", alias, path)
		}
		fmt.Println(color.GreenString("Password store %s initialized for:", path))
		for _, r := range s.Recipients() {
			color.Yellow(r)
		}
	}

	r.mounts[alias] = s
	return nil
}

// RemoveMount removes and existing mount
func (r *Store) RemoveMount(alias string) error {
	if _, found := r.mounts[alias]; !found {
		return errors.Errorf("%s is not mounted", alias)
	}
	if _, found := r.mounts[alias]; !found {
		fmt.Println(color.YellowString("%s is not initialized", alias))
	}
	delete(r.mounts, alias)
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
func (r *Store) getStore(name string) *sub.Store {
	name = strings.TrimSuffix(name, "/")
	mp := r.mountPoint(name)
	if sub, found := r.mounts[mp]; found {
		return sub
	}
	return r.store
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
