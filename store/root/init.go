package root

import (
	"context"

	"github.com/justwatchcom/gopass/store/sub"
	"github.com/justwatchcom/gopass/utils/fsutil"
	"github.com/pkg/errors"
)

// Initialized checks on disk if .gpg-id was generated and thus returns true.
func (r *Store) Initialized() bool {
	return r.store.Initialized()
}

// Init tries to initalize a new password store location matching the object
func (r *Store) Init(ctx context.Context, alias, path string, ids ...string) error {
	cfg := r.Config()
	cfg.Path = fsutil.CleanPath(path)
	sub, err := sub.New(alias, cfg)
	if err != nil {
		return errors.Wrapf(err, "failed to create new sub store '%s'", alias)
	}
	if !r.store.Initialized() && alias == "" {
		r.store = sub
	}

	return sub.Init(ctx, path, ids...)
}
