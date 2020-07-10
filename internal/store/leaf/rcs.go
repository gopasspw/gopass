package leaf

import (
	"context"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/backend/storage/kv/ondisk"
	"github.com/gopasspw/gopass/internal/debug"
)

func (s *Store) initRCSBackend(ctx context.Context) {
	if rcs, ok := s.storage.(*ondisk.OnDisk); ok {
		s.rcs = rcs
		return
	}
	rcs, err := backend.DetectRCS(ctx, s.path)
	if err != nil {
		debug.Log("Failed to initialized RCS backend: %s", err)
	}
	s.rcs = rcs
}
