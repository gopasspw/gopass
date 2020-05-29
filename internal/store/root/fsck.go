package root

import (
	"context"
	"strings"

	"github.com/gopasspw/gopass/internal/debug"
	"github.com/gopasspw/gopass/internal/out"
	multierror "github.com/hashicorp/go-multierror"
)

// Fsck checks all stores/entries matching the given prefix
func (s *Store) Fsck(ctx context.Context, path string) error {
	var result error

	for alias, sub := range s.mounts {
		if sub == nil {
			continue
		}
		if path != "" && !strings.HasPrefix(path, alias+"/") {
			continue
		}
		path = strings.TrimPrefix(path, alias+"/")
		debug.Log("root.Fsck() - Checking %s", alias)
		if err := sub.Fsck(ctx, path); err != nil {
			out.Error(ctx, "fsck failed on sub store %s: %s", alias, err)
			result = multierror.Append(result, err)
		}
	}
	if err := s.store.Fsck(ctx, path); err != nil {
		out.Error(ctx, "fsck failed on root store: %s", err)
		result = multierror.Append(result, err)
	}

	return result
}
