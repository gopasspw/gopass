package sub

import (
	"context"

	"github.com/justwatchcom/gopass/pkg/out"
	"github.com/justwatchcom/gopass/pkg/store"
	"github.com/justwatchcom/gopass/pkg/store/secret"
)

// Get returns the plaintext of a single key
func (s *Store) Get(ctx context.Context, name string) (store.Secret, error) {
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
