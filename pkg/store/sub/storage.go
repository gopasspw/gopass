package sub

import (
	"context"

	"github.com/gopasspw/gopass/pkg/backend"
	"github.com/gopasspw/gopass/pkg/ctxutil"

	"github.com/pkg/errors"
)

func (s *Store) initStorageBackend(ctx context.Context) error {
	ctx = ctxutil.WithAlias(ctx, s.alias)
	store, err := backend.NewStorage(ctx, s.url.Storage, s.url)
	if err != nil {
		return errors.Wrapf(err, "unknown storage backend")
	}
	s.storage = store
	return nil
}
