package root

import "strings"

// GetKey will return a single named entry from a structured document (YAML)
// in secret name. If no such key exists or yaml decoding fails it will
// return an error
func (r *Store) GetKey(name, key string) ([]byte, error) {
	store := r.getStore(name)
	return store.GetKey(strings.TrimPrefix(name, store.Alias()), key)
}

// SetKey sets a single key in structured document (YAML) to the given
// value. If the secret name is non-empty but no YAML it will return an error.
func (r *Store) SetKey(name, key, value string) error {
	store := r.getStore(name)
	return store.SetKey(strings.TrimPrefix(name, store.Alias()), key, value)
}

// DeleteKey removes a single key
func (r *Store) DeleteKey(name, key string) error {
	store := r.getStore(name)
	return store.DeleteKey(strings.TrimPrefix(name, store.Alias()), key)
}
