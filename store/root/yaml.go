package root

import (
	"context"
	"strings"
)

// GetKey will return a single named entry from a structured document (YAML)
// in secret name. If no such key exists or yaml decoding fails it will
// return an error
func (r *Store) GetKey(ctx context.Context, name, key string) ([]byte, error) {
	store := r.getStore(name)
	return store.GetKey(ctx, strings.TrimPrefix(name, store.Alias()), key)
}

// SetKey sets a single key in structured document (YAML) to the given
// value. If the secret name is non-empty but no YAML it will return an error.
func (r *Store) SetKey(ctx context.Context, name, key, value string) error {
	store := r.getStore(name)
	return store.SetKey(ctx, strings.TrimPrefix(name, store.Alias()), key, value)
}

// DeleteKey removes a single key
func (r *Store) DeleteKey(ctx context.Context, name, key string) error {
	store := r.getStore(name)
	return store.DeleteKey(ctx, strings.TrimPrefix(name, store.Alias()), key)
}
