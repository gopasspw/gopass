package binary

import (
	"context"

	"github.com/gopasspw/gopass/internal/store"
)

const (
	// Suffix is the suffix that is appended to binaries in the store
	Suffix = ".b64"
)

type storer interface {
	Get(context.Context, string) (store.Secret, error)
	Set(context.Context, string, store.Secret) error
	Exists(context.Context, string) bool
	Delete(context.Context, string) error
}
