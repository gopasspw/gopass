// Package api provides a gopass API implementation.
// It provides a simple interface to interact with the gopass password store.
// It is used by the gopass CLI and other tools to access the password store.
package api

import (
	"context"
	"fmt"

	// load crypto backends.
	_ "github.com/gopasspw/gopass/internal/backend/crypto"
	// load storage backends.
	_ "github.com/gopasspw/gopass/internal/backend/storage"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/queue"
	"github.com/gopasspw/gopass/internal/store/root"
	"github.com/gopasspw/gopass/internal/tree"
	"github.com/gopasspw/gopass/pkg/gopass"
)

// Gopass is a secret store implementation. It is the main entry point for the gopass API.
type Gopass struct {
	rs *root.Store
}

// make sure that *Gopass implements Store.
var _ gopass.Store = &Gopass{}

// ErrNotImplemented is returned when a method is not yet implemented.
var ErrNotImplemented = fmt.Errorf("not yet implemented")

// ErrNotInitialized is returned when the store is not initialized.
// This happens when the gopass CLI has not been used to set up a password store yet.
var ErrNotInitialized = fmt.Errorf("password store not initialized. run 'gopass setup' first")

// New initializes an existing password store and returns a new Gopass instance.
// It will attempt to load an existing configuration or use the built-in defaults.
// If no password store is found the user will need to initialize it with the gopass
// CLI (`gopass setup`) first.
//
// WARNING: This will need to change to accommodate for runtime configuration.
func New(ctx context.Context) (*Gopass, error) {
	cfg := config.New()
	store := root.New(cfg)

	initialized, err := store.IsInitialized(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check initialization: %w", err)
	}

	if !initialized {
		return nil, ErrNotInitialized
	}

	return &Gopass{
		rs: store,
	}, nil
}

// List returns a list of all secret names.
func (g *Gopass) List(ctx context.Context) ([]string, error) {
	return g.rs.List(ctx, tree.INF) //nolint:wrapcheck
}

// Get returns a single, encrypted secret. It must be unwrapped before use.
// Use "latest" to get the latest revision.
// The revision parameter is not yet implemented.
func (g *Gopass) Get(ctx context.Context, name, revision string) (gopass.Secret, error) {
	return g.rs.Get(ctx, name) //nolint:wrapcheck
}

// Set adds a new revision to an existing secret or creates a new one if it doesn't exist.
// Create new secrets with secrets.New().
// The secret is passed as a gopass.Byter, which is an interface that can be satisfied by a []byte.
func (g *Gopass) Set(ctx context.Context, name string, sec gopass.Byter) error {
	return g.rs.Set(ctx, name, sec) //nolint:wrapcheck
}

// Remove removes a single secret.
func (g *Gopass) Remove(ctx context.Context, name string) error {
	return g.rs.Delete(ctx, name) //nolint:wrapcheck
}

// RemoveAll removes all secrets with a given prefix.
func (g *Gopass) RemoveAll(ctx context.Context, prefix string) error {
	return g.rs.Prune(ctx, prefix) //nolint:wrapcheck
}

// Rename moves a secret from one path to another.
func (g *Gopass) Rename(ctx context.Context, src, dest string) error {
	return g.rs.Move(ctx, src, dest) //nolint:wrapcheck
}

// Sync synchronizes the store with a remote.
// Not yet implemented.
func (g *Gopass) Sync(ctx context.Context) error {
	return ErrNotImplemented
}

// Revisions lists all revisions of this secret.
// Not yet implemented.
func (g *Gopass) Revisions(ctx context.Context, name string) ([]string, error) {
	return nil, ErrNotImplemented
}

// String returns the name of this store.
func (g *Gopass) String() string {
	return "gopass"
}

// Close shuts down all background processes and closes the store.
//
// MUST be called before exiting to make sure any background processing
// (e.g. pending commits or pushes) are complete. Failing to do so might
// result in an invalid password store state.
func (g *Gopass) Close(ctx context.Context) error {
	if err := queue.GetQueue(ctx).Close(ctx); err != nil {
		return fmt.Errorf("failed to close queue: %w", err)
	}

	return nil
}

// ConfigDir returns gopass' configuration directory.
func ConfigDir() string {
	return config.Directory()
}
