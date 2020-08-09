package api

import (
	"context"
	"fmt"

	_ "github.com/gopasspw/gopass/internal/backend/crypto"  // load crypto backends
	_ "github.com/gopasspw/gopass/internal/backend/storage" // load storage backends
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/queue"
	"github.com/gopasspw/gopass/internal/store/root"
	"github.com/gopasspw/gopass/pkg/gopass"
)

// Gopass is a secret store implementation
type Gopass struct {
	rs *root.Store
}

// make sure that *Gopass implements Store
var _ gopass.Store = &Gopass{}

// New creates a new secret store
// WARNING: This will need to change to accommodate for runtime configuration.
func New(ctx context.Context) (*Gopass, error) {
	cfg := config.LoadWithFallback() // TODO(dominikschulz) Consider removing WithFallback
	store := root.New(cfg)
	initialized, err := store.Initialized(ctx)
	if err != nil {
		return nil, err
	}
	if !initialized {
		return nil, fmt.Errorf("store not initialized. run gopass init first")
	}
	return &Gopass{
		rs: store,
	}, nil
}

// List returns a list of all secrets.
func (g *Gopass) List(ctx context.Context) ([]string, error) {
	return g.rs.List(ctx, 0)
}

// Get returns a single, encrypted secret. It must be unwrapped before use.
func (g *Gopass) Get(ctx context.Context, name, revision string) (gopass.Secret, error) {
	return g.rs.Get(ctx, name)
}

// Set adds a new revision to an existing secret or creates a new one.
func (g *Gopass) Set(ctx context.Context, name string, sec gopass.Byter) error {
	return g.rs.Set(ctx, name, sec)
}

// Remove removes a single secret.
func (g *Gopass) Remove(ctx context.Context, name string) error {
	return g.rs.Delete(ctx, name)
}

// RemoveAll removes all secrets with a given prefix.
func (g *Gopass) RemoveAll(ctx context.Context, prefix string) error {
	return g.rs.Prune(ctx, prefix)
}

// Rename move a prefix to another.
func (g *Gopass) Rename(ctx context.Context, src, dest string) error {
	return g.rs.Move(ctx, src, dest)
}

// Sync synchronizes a secret with a remote
func (g *Gopass) Sync(ctx context.Context) error {
	return fmt.Errorf("not yet implemented")
}

// Revisions lists all revisions of this secret
func (g *Gopass) Revisions(ctx context.Context, name string) ([]string, error) {
	return nil, fmt.Errorf("not yet implemented")
}

func (g *Gopass) String() string {
	return "gopass"
}

// Close shuts down all background processes
func (g *Gopass) Close(ctx context.Context) error {
	return queue.GetQueue(ctx).Wait(ctx)
}
