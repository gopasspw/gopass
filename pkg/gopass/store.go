// Package gopass provides the interfaces for the gopass password store.
package gopass

import (
	"context"
	"fmt"
)

// Byter is a minimal secrets write interface. It is used for writing secrets to the store.
type Byter interface {
	Bytes() []byte
}

// Secret is a secret type that represents a single secret in the store.
type Secret interface {
	Byter

	// Keys returns a list of all keys in the secret's metadata.
	Keys() []string
	// Get returns a single value for a given key from the secret's metadata.
	Get(key string) (string, bool)
	// Values returns all values for a given key from the secret's metadata.
	Values(key string) ([]string, bool)
	// Set sets a single value for a given key in the secret's metadata.
	Set(key string, value any) error
	// Add adds a new value for a given key in the secret's metadata.
	Add(key string, value any) error
	// Del removes a key from the secret's metadata.
	Del(key string) bool

	// Ref returns a reference to another secret if the secret is a reference.
	// A reference is a secret that contains a gopass:// URI.
	Ref() (string, bool)

	// Body returns the body of the secret, which is everything except the password and metadata.
	Body() string
	// Password returns the password of the secret.
	Password() string
	// SetPassword sets the password of the secret.
	SetPassword(string)
}

// Store is a secret store that provides methods for interacting with the password store.
type Store interface {
	fmt.Stringer

	// List lists all secrets in the store.
	List(context.Context) ([]string, error)
	// Get returns a decrypted secret from the store.
	// The revision parameter defaults to "latest".
	Get(ctx context.Context, name, revision string) (Secret, error)
	// Set creates or updates a secret in the store.
	Set(ctx context.Context, name string, sec Byter) error
	// Revisions returns a list of all revisions of a secret.
	Revisions(ctx context.Context, name string) ([]string, error)
	// Remove removes a single secret from the store.
	Remove(ctx context.Context, name string) error
	// RemoveAll removes all secrets with a common prefix from the store.
	RemoveAll(ctx context.Context, prefix string) error
	// Rename renames a secret or a folder of secrets.
	Rename(ctx context.Context, src, dest string) error
	// Sync synchronizes the store with a remote.
	// NOTE: The store is automatically synchronized when mutated.
	// This method can be used to manually trigger a synchronization.
	Sync(ctx context.Context) error
	// Close closes the store and cleans up any resources.
	// It MUST be called before the program exits.
	Close(ctx context.Context) error
}
