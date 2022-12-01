package leaf

import (
	"context"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/gopasspw/gopass/pkg/gopass/secrets/secparse"
)

// Get returns the plaintext of a single key.
func (s *Store) Get(ctx context.Context, name string) (gopass.Secret, error) {
	p := s.Passfile(name)

	ciphertext, err := s.storage.Get(ctx, p)
	if err != nil {
		debug.Log("File %s not found: %s", p, err)

		return nil, store.ErrNotFound
	}

	content, err := s.crypto.Decrypt(ctx, ciphertext)
	if err != nil {
		out.Errorf(ctx, "Decryption failed: %s\n%s", err, string(content))

		return nil, store.ErrDecrypt
	}

	if !ctxutil.IsShowParsing(ctx) {
		debug.Log("secrets parsing is disabled. parsing as AKV")

		return secrets.ParseAKV(content), nil
	}

	debug.Log("secrets parsing is enabled")

	return secparse.Parse(content)
}
