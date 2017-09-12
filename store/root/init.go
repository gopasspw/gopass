package root

import (
	"context"

	"github.com/justwatchcom/gopass/store/sub"
)

// Initialized checks on disk if .gpg-id was generated and thus returns true.
func (r *Store) Initialized() bool {
	return r.store.Initialized()
}

// Init tries to initalize a new password store location matching the object
func (r *Store) Init(ctx context.Context, alias, path string, ids ...string) error {
	sub := sub.New(alias, path)
	if !r.store.Initialized() && alias == "" {
		r.store = sub
	}

	return sub.Init(ctx, path, ids...)
}
