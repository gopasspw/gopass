package gopass

import (
	"context"
	"fmt"
)

// Byter is a minimal secrets write interface
type Byter interface {
	Bytes() []byte
}

// Secret is a secret type.
type Secret interface {
	Byter

	Keys() []string
	// Get returns a single header value, use Password() to get the password value.
	Get(key string) (string, bool)
	// Set sets a single header value, use SetPassword() to set the password value.
	Set(key string, value interface{}) error
	// Del removes a single header value
	Del(key string) bool

	// GetBody returns everything except the header. Use Bytes to get everything
	Body() string
	Password() string
	SetPassword(string)
}

// Store is a secret store.
type Store interface {
	fmt.Stringer

	// List all secrets
	List(context.Context) ([]string, error)
	// Get an decrypted secret. Revision defaults to "latest".
	Get(ctx context.Context, name, revision string) (Secret, error)
	// Set (add) a new revision of an secret
	Set(ctx context.Context, name string, sec Byter) error
	// Revisions is TODO
	Revisions(ctx context.Context, name string) ([]string, error)
	// Remove a single secret
	Remove(ctx context.Context, name string) error
	// RemoveAll secrets with a common prefix
	RemoveAll(ctx context.Context, prefix string) error
	// Rename a path (secret of prefix) without decrypting
	Rename(ctx context.Context, src, dest string) error
	// Sync with a remote (if configured)
	// NOTE: We will always auto-sync when mutating the store. Use this to
	// manually pull in changes.
	Sync(ctx context.Context) error
	// Clean up any resources. MUST be called before the process exists.
	Close(ctx context.Context) error
}
