package leaf

import (
	"context"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/debug"
)

func (s *Store) initRCSBackend(ctx context.Context) {
	rcs, err := backend.DetectRCS(ctx, s.path)
	if err != nil {
		debug.Log("Failed to initialized RCS backend: %s", err)
	}
	s.rcs = rcs
}
