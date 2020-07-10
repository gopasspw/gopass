package gopass

import (
	"context"
	"fmt"

	"github.com/gopasspw/gopass/pkg/gopass/secret"
)

// Byter is a minimal secrets write interface
type Byter interface {
	Bytes() []byte
}

// Secret is a secret type.
type Secret interface {
	Byter

	Keys() []string
	// Get returns a single header value, use Get("Password") to get the
	// password value.
	Get(key string) string
	// Set sets a single header value, use Set("Password") to set the password
	// value.
	Set(key, value string)
	// Del removes a single header value
	Del(key string)

	// GetBody returns everything except the header. Use Bytes to get everything
	GetBody() string

	// MIME converts the secret to a MIME secret
	MIME() *secret.MIME
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
	Close(ctx context.Context)
}
