package root

import (
	"strings"

	"github.com/justwatchcom/gopass/store"
)

// Set encodes and write the ciphertext of one entry to disk
func (r *Store) Set(name string, content []byte, reason string) error {
	store := r.getStore(name)
	return store.Set(strings.TrimPrefix(name, store.Alias()), content, reason)
}

// SetPassword Update only the first line in an already existing entry
func (r *Store) SetPassword(name string, password []byte) error {
	store := r.getStore(name)
	return store.SetPassword(strings.TrimPrefix(name, store.Alias()), password)
}

// SetConfirm calls Set with confirmation callback
func (r *Store) SetConfirm(name string, content []byte, reason string, cb store.RecipientCallback) error {
	store := r.getStore(name)
	return store.SetConfirm(strings.TrimPrefix(name, store.Alias()), content, reason, cb)
}
