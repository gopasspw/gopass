package sub

import (
	"context"

	"github.com/gopasspw/gopass/pkg/backend"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/pkg/errors"
)

func (s *Store) initRCSBackend(ctx context.Context) error {
	rcs, err := backend.OpenRCS(ctx, s.url.RCS, s.url.Path)
	if err != nil {
		if errors.Cause(err) == backend.ErrNotFound {
			return err
		}
		out.Debug(ctx, "Failed to initialized RCS backend: %s", err)
		return nil
	}
	s.rcs = rcs
	return nil
}
