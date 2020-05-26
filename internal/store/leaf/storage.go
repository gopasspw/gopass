package leaf

import (
	"context"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/pkg/ctxutil"

	"github.com/pkg/errors"
)

func (s *Store) initStorageBackend(ctx context.Context) error {
	ctx = ctxutil.WithAlias(ctx, s.alias)
	store, err := backend.DetectStorage(ctx, s.path)
	if err != nil {
		return errors.Wrapf(err, "unknown storage backend")
	}
	s.storage = store
	return nil
}
