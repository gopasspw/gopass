package root

import (
	"context"
	"strings"

	"github.com/pkg/errors"
)

// Fsck checks all stores/entries matching the given prefix
func (s *Store) Fsck(ctx context.Context, path string) error {
	for alias, sub := range s.mounts {
		if sub == nil {
			continue
		}
		if path != "" && !strings.HasPrefix(path, alias+"/") {
			continue
		}
		if err := sub.Fsck(ctx, strings.TrimPrefix(path, alias+"/")); err != nil {
			return errors.Wrapf(err, "fsck failed on sub store %s: %s", alias, err)
		}
	}
	if err := s.store.Fsck(ctx, path); err != nil {
		return errors.Wrapf(err, "fsck failed on root store: %s", err)
	}
	return nil
}
