package sub

import (
	"context"
	"fmt"
	"strings"

	"github.com/justwatchcom/gopass/store"
	"github.com/justwatchcom/gopass/store/secret"
	"github.com/justwatchcom/gopass/utils/fsutil"
)

// Get returns the plaintext of a single key
func (s *Store) Get(ctx context.Context, name string) (*secret.Secret, error) {
	p := s.passfile(name)

	if !strings.HasPrefix(p, s.path) {
		return nil, store.ErrSneaky
	}

	if !fsutil.IsFile(p) {
		if s.debug {
			fmt.Printf("File %s not found\n", p)
		}
		return nil, store.ErrNotFound
	}

	content, err := s.gpg.Decrypt(ctx, p)
	if err != nil {
		return nil, store.ErrDecrypt
	}

	return secret.Parse(content)
}
