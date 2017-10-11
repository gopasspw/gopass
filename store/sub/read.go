package sub

import (
	"context"
	"strings"

	"github.com/justwatchcom/gopass/store"
	"github.com/justwatchcom/gopass/store/secret"
	"github.com/justwatchcom/gopass/utils/fsutil"
	"github.com/justwatchcom/gopass/utils/out"
)

// Get returns the plaintext of a single key
func (s *Store) Get(ctx context.Context, name string) (*secret.Secret, error) {
	p := s.passfile(name)

	if !strings.HasPrefix(p, s.path) {
		return nil, store.ErrSneaky
	}

	if !fsutil.IsFile(p) {
		out.Debug(ctx, "File %s not found", p)
		return nil, store.ErrNotFound
	}

	content, err := s.gpg.Decrypt(ctx, p)
	if err != nil {
		return nil, store.ErrDecrypt
	}

	sec, err := secret.Parse(content)
	if err != nil {
		out.Debug(ctx, "Failed to parse YAML: %s", err)
	}
	return sec, nil
}
