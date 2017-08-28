package root

import "strings"

// Get returns the plaintext of a single key
func (r *Store) Get(name string) ([]byte, error) {
	// forward to substore
	store := r.getStore(name)
	return store.Get(strings.TrimPrefix(name, store.Alias()))
}

// GetFirstLine returns the first line of the plaintext of a single key
func (r *Store) GetFirstLine(name string) ([]byte, error) {
	store := r.getStore(name)
	return store.GetFirstLine(strings.TrimPrefix(name, store.Alias()))
}

// GetBody returns everything but the first line from a key
func (r *Store) GetBody(name string) ([]byte, error) {
	store := r.getStore(name)
	return store.GetBody(strings.TrimPrefix(name, store.Alias()))
}
