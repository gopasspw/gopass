package leaf

import (
	"context"
	"path/filepath"
	"strings"
)

type linkTargeter interface {
	LinkTarget(context.Context, string) (string, bool, error)
}

// LinkTarget returns the target of a linked secret, if the named entry is a link.
func (s *Store) LinkTarget(ctx context.Context, name string) (string, bool, error) {
	lt, ok := s.storage.(linkTargeter)
	if !ok {
		return "", false, nil
	}

	if s.alias != "" {
		name = strings.TrimPrefix(name, s.alias+Sep)
	}

	target, isLink, err := lt.LinkTarget(ctx, s.passfile(ctx, name))
	if err != nil || !isLink {
		return "", isLink, err
	}

	target = strings.TrimSuffix(filepath.ToSlash(target), "."+s.crypto.Ext())
	if s.alias != "" {
		target = s.alias + Sep + target
	}

	return target, true, nil
}
