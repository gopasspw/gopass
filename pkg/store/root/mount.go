package root

import (
	"context"
	"sort"
	"strings"

	"github.com/gopasspw/gopass/pkg/backend"
	"github.com/gopasspw/gopass/pkg/config"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/pkg/store"
	"github.com/gopasspw/gopass/pkg/store/sub"

	"github.com/fatih/color"
	"github.com/pkg/errors"
)

var sep = "/"

// AddMount adds a new mount
func (r *Store) AddMount(ctx context.Context, alias, path string, keys ...string) error {
	if err := r.addMount(ctx, alias, path, nil, keys...); err != nil {
		return errors.Wrapf(err, "failed to add mount")
	}

	// check for duplicate mounts
	return r.checkMounts()
}

func (r *Store) addMount(ctx context.Context, alias, path string, sc *config.StoreConfig, keys ...string) error {
	if alias == "" {
		return errors.Errorf("alias must not be empty")
	}
	if r.mounts == nil {
		r.mounts = make(map[string]store.Store, 1)
	}
	if _, found := r.mounts[alias]; found {
		return AlreadyMountedError(alias)
	}

	out.Debug(ctx, "addMount - Path: %s - StoreConfig: %+v", path, sc)
	// propagate our config settings to the sub store
	if sc != nil {
		if !backend.HasCryptoBackend(ctx) {
			ctx = backend.WithCryptoBackend(ctx, sc.Path.Crypto)
			out.Debug(ctx, "addMount - Using crypto backend %s", backend.CryptoBackendName(sc.Path.Crypto))
		}
		if !backend.HasRCSBackend(ctx) {
			ctx = backend.WithRCSBackend(ctx, sc.Path.RCS)
			out.Debug(ctx, "addMount - Using RCS backend %s", backend.RCSBackendName(sc.Path.RCS))
		}
	}

	// parse backend URL
	pathURL, err := backend.ParseURL(path)
	if err != nil {
		return errors.Wrapf(err, "failed to parse backend URL '%s': %s", path, err)
	}

	// initialize sub store
	s, err := r.initSub(ctx, sc, alias, pathURL, keys)
	if err != nil {
		return errors.Wrapf(err, "failed to init sub store '%s' at '%s'", alias, pathURL)
	}

	r.mounts[alias] = s
	if r.cfg.Mounts == nil {
		r.cfg.Mounts = make(map[string]*config.StoreConfig, 1)
	}
	if sc == nil {
		// imporant: copy root config to avoid overwriting it with sub store
		// values
		cp := *r.cfg.Root
		sc = &cp
	}
	sc.Path = pathURL
	if backend.HasCryptoBackend(ctx) {
		sc.Path.Crypto = backend.GetCryptoBackend(ctx)
	}
	if backend.HasRCSBackend(ctx) {
		sc.Path.RCS = backend.GetRCSBackend(ctx)
	}
	if backend.HasStorageBackend(ctx) {
		sc.Path.Storage = backend.GetStorageBackend(ctx)
	}
	r.cfg.Mounts[alias] = sc

	out.Debug(ctx, "Added mount %s -> %s", alias, sc.Path.String())
	return nil
}

func (r *Store) initSub(ctx context.Context, sc *config.StoreConfig, alias string, path *backend.URL, keys []string) (store.Store, error) {
	// init regular sub store
	s, err := sub.New(ctx, r.cfg, alias, path, r.cfg.Directory())
	if err != nil {
		return nil, errors.Wrapf(err, "failed to initialize store '%s' at '%s': %s", alias, path, err)
	}

	if s.Initialized(ctx) {
		return s, nil
	}

	out.Debug(ctx, "[%s] Mount %s is not initialized", alias, path)
	if len(keys) < 1 {
		return s, NotInitializedError{alias, path.String()}
	}
	if err := s.Init(ctx, path.String(), keys...); err != nil {
		return s, errors.Wrapf(err, "failed to initialize store '%s' at '%s'", alias, path)
	}
	out.Green(ctx, "Password store %s initialized for:", path)
	for _, r := range s.Recipients(ctx) {
		color.Yellow(r)
	}

	return s, nil
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

// MountPoint returns the most-specific mount point for the given key
func (r *Store) MountPoint(name string) string {
	for _, mp := range r.MountPoints() {
		if strings.HasPrefix(name+sep, mp+sep) {
			return mp
		}
	}
	return ""
}

// getStore returns the Store object at the most-specific mount point for the
// given key
// context with sub store options set, sub store reference, truncated path to secret
func (r *Store) getStore(ctx context.Context, name string) (context.Context, store.Store, string) {
	name = strings.TrimSuffix(name, sep)
	mp := r.MountPoint(name)
	if sub, found := r.mounts[mp]; found {
		return r.cfg.Mounts[mp].WithContext(ctx), sub, strings.TrimPrefix(name, sub.Alias())
	}
	return r.cfg.Root.WithContext(ctx), r.store, name
}

// WithConfig populates the context with the substore config
func (r *Store) WithConfig(ctx context.Context, name string) context.Context {
	name = strings.TrimSuffix(name, sep)
	mp := r.MountPoint(name)
	if _, found := r.mounts[mp]; found {
		return r.cfg.Mounts[mp].WithContext(ctx)
	}
	return r.cfg.Root.WithContext(ctx)
}

// GetSubStore returns an exact match for a mount point or an error if this
// mount point does not exist
func (r *Store) GetSubStore(name string) (store.Store, error) {
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
