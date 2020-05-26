package leaf

import (
	"context"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/out"
)

func (s *Store) initRCSBackend(ctx context.Context) error {
	rcs, err := backend.DetectRCS(ctx, s.path)
	if err != nil {
		out.Debug(ctx, "Failed to initialized RCS backend: %s", err)
		return nil
	}
	s.rcs = rcs
	return nil
}
