package sub

import (
	"context"
	"path/filepath"
	"strings"
)

var (
	sep = string(filepath.Separator)
)

// List will list all entries in this store
func (s *Store) List(ctx context.Context, prefix string) ([]string, error) {
	lst, err := s.store.List(ctx, prefix)
	if err != nil {
		return nil, err
	}
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
