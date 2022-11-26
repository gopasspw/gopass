package leaf

import (
	"context"
	"strings"

	"github.com/gopasspw/gopass/pkg/debug"
)

// Sep is the separator used in lists to separate folders from entries.
var Sep = "/"

// List will list all entries in this store.
func (s *Store) List(ctx context.Context, prefix string) ([]string, error) {
	if s.storage == nil || s.crypto == nil {
		return nil, nil
	}

	lst, err := s.storage.List(ctx, prefix)
	if err != nil {
		return nil, err
	}

	debug.Log("Listing storage content of %s: %+v", prefix, lst)
	out := make([]string, 0, len(lst))
	cExt := "." + s.crypto.Ext()
	for _, path := range lst {
		if !strings.HasSuffix(path, cExt) {
			continue
		}
		path = strings.TrimSuffix(path, cExt)
		if s.alias != "" {
			path = s.alias + Sep + path
		}
		out = append(out, path)
	}
	debug.Log("Leaf store entries: %+v", out)

	return out, nil
}
