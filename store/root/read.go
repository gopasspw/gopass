package root

import (
	"context"
	"strings"
)

// Get returns the plaintext of a single key
func (r *Store) Get(ctx context.Context, name string) ([]byte, error) {
	// forward to substore
	store := r.getStore(name)
	return store.Get(ctx, strings.TrimPrefix(name, store.Alias()))
}

// GetFirstLine returns the first line of the plaintext of a single key
func (r *Store) GetFirstLine(ctx context.Context, name string) ([]byte, error) {
	store := r.getStore(name)
	return store.GetFirstLine(ctx, strings.TrimPrefix(name, store.Alias()))
}

// GetBody returns everything but the first line from a key
func (r *Store) GetBody(ctx context.Context, name string) ([]byte, error) {
	store := r.getStore(name)
	return store.GetBody(ctx, strings.TrimPrefix(name, store.Alias()))
}
