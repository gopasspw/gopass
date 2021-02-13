package leaf

import (
	"context"

	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/gopasspw/gopass/pkg/gopass/secrets/secparse"
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
		out.Errorf(ctx, "Decryption failed: %s\n%s", err, string(content))
		return nil, store.ErrDecrypt
	}
	var sec gopass.Secret
	sec = secrets.ParsePlain(content)
	if ctxutil.IsShowParsing(ctx) {
		sec, err = secparse.Parse(content)
		if err != nil {
			debug.Log("Failed to parse secret: %s", err)
		}
	}

	return sec, nil
}
