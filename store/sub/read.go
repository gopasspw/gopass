package sub

import (
	"context"

	"github.com/justwatchcom/gopass/store"
	"github.com/justwatchcom/gopass/store/secret"
	"github.com/justwatchcom/gopass/utils/out"
)

// Get returns the plaintext of a single key
func (s *Store) Get(ctx context.Context, name string) (*secret.Secret, error) {
	p := s.passfile(name)

	ciphertext, err := s.storage.Get(ctx, p)
	if err != nil {
		out.Debug(ctx, "File %s not found: %s", p, err)
		return nil, store.ErrNotFound
	}

	content, err := s.crypto.Decrypt(ctx, ciphertext)
	if err != nil {
		out.Debug(ctx, "Decryption failed: %s", err)
		return nil, store.ErrDecrypt
	}

	sec, err := secret.Parse(content)
	if err != nil {
		out.Debug(ctx, "Failed to parse YAML: %s", err)
	}
	return sec, nil
}
