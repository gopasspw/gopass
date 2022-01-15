package root

import (
	"context"
	"fmt"
	"math"
	"runtime"
	"strings"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/store/leaf"
	"github.com/gopasspw/gopass/pkg/debug"
)

// Store is the public facing password store. It contains one or more
// leaf stores.
type Store struct {
	cfg    *config.Config
	mounts map[string]*leaf.Store
	store  *leaf.Store
}

// New creates a new store.
func New(cfg *config.Config) *Store {
	if cfg == nil {
		cfg = &config.Config{}
	}
	r := &Store{
		cfg:    cfg,
		mounts: make(map[string]*leaf.Store, len(cfg.Mounts)),
	}

	debug.Log("created store %s", r)
	return r
}

// WithContext populates the context with the store config.
func (r *Store) WithContext(ctx context.Context) context.Context {
	return r.cfg.WithContext(ctx)
}

// Exists checks the existence of a single entry.
func (r *Store) Exists(ctx context.Context, name string) bool {
	store, name := r.getStore(name)
	return store.Exists(ctx, name)
}

// IsDir checks if a given key is actually a folder.
func (r *Store) IsDir(ctx context.Context, name string) bool {
	store, name := r.getStore(name)
	return store.IsDir(ctx, name)
}

func (r *Store) String() string {
	ms := make([]string, 0, len(r.mounts))
	for alias, sub := range r.mounts {
		ms = append(ms, alias+"="+sub.String())
	}
	path := ""
	if r.store != nil {
		path = r.store.Path()
	}
	return fmt.Sprintf("Store(Path: %s, Mounts: %+v)", path, strings.Join(ms, ","))
}

// Path returns the store path.
func (r *Store) Path() string {
	if r.store == nil {
		return ""
	}
	return r.store.Path()
}

// Alias always returns an empty string.
func (r *Store) Alias() string {
	return ""
}

// Storage returns the storage backend for the given mount point.
func (r *Store) Storage(ctx context.Context, name string) backend.Storage {
	sub, _ := r.getStore(name)
	if sub == nil || !sub.Valid() {
		return nil
	}
	return sub.Storage()
}

// Concurrency returns the concurrency level supported by this store,
// which is the minimum of all mount points.
func (r *Store) Concurrency() int {
	min := math.MaxInt
	for _, sub := range r.mounts {
		if sub.Concurrency() < min {
			min = sub.Concurrency()
		}
	}
	if nc := runtime.NumCPU(); nc < min {
		min = nc
	}
	return min
}
