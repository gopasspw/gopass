package root

import (
	"context"
	"strings"

	"github.com/gopasspw/gopass/pkg/out"
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
		path = strings.TrimPrefix(path, alias+"/")
		out.Debug(ctx, "root.Fsck() - Checking %s", alias)
		if err := sub.Fsck(ctx, path); err != nil {
			return errors.Wrapf(err, "fsck failed on sub store %s: %s", alias, err)
		}
	}
	if err := s.store.Fsck(ctx, path); err != nil {
		return errors.Wrapf(err, "fsck failed on root store: %s", err)
	}
	return nil
}
