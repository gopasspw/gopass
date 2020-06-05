package leaf

import (
	"context"

	"github.com/gopasspw/gopass/internal/debug"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/gopasspw/gopass/pkg/gopass/secret/secparse"
)

// Get returns the plaintext of a single key
func (s *Store) Get(ctx context.Context, name string) (gopass.Secret, error) {
	p := s.passfile(name)

	ciphertext, err := s.storage.Get(ctx, p)
	if err != nil {
		debug.Log("File %s not found: %s", p, err)
		return nil, store.ErrNotFound
	}

	content, err := s.crypto.Decrypt(ctx, ciphertext)
	if err != nil {
		out.Error(ctx, "Decryption failed: %s\n%s", err, string(content))
		return nil, store.ErrDecrypt
	}

	sec, err := secparse.Parse(content)
	if err != nil {
		debug.Log("Failed to parse secret: %s", err)
	}
	return sec, nil
}
