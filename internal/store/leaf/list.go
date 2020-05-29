package leaf

import (
	"context"
	"strings"

	"github.com/gopasspw/gopass/internal/debug"
)

var (
	sep = "/"
)

// List will list all entries in this store
func (s *Store) List(ctx context.Context, prefix string) ([]string, error) {
	if s.storage == nil || s.crypto == nil {
		return nil, nil
	}

	lst, err := s.storage.List(ctx, prefix)
	if err != nil {
		return nil, err
	}
	debug.Log("sub.List(%s): %+v\n", prefix, lst)
	out := make([]string, 0, len(lst))
	cExt := "." + s.crypto.Ext()
	for _, path := range lst {
		if !strings.HasSuffix(path, cExt) {
			continue
		}
		path = strings.TrimSuffix(path, cExt)
		if s.alias != "" {
			path = s.alias + sep + path
		}
		out = append(out, path)
	}
	return out, nil
}
