package sub

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/gopasspw/gopass/pkg/out"
)

var (
	sep = string(filepath.Separator)
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
	out.Debug(ctx, "sub.List(%s): %+v\n", prefix, lst)
	out := make([]string, 0, len(lst))
	cExt := "." + s.crypto.Ext()
	for _, path := range lst {
		if !strings.HasSuffix(path, cExt) {
			continue
		}
		path = strings.TrimSuffix(path, cExt)
		if s.alias != "" {
			path = s.alias + "/" + path
		}
		out = append(out, path)
	}
	return out, nil
}
