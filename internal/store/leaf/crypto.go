package leaf

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/backend/crypto/age"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/pkg/errors"
)

func (s *Store) initCryptoBackend(ctx context.Context) error {
	cb, err := backend.DetectCrypto(ctx, s.storage)
	if err != nil {
		return err
	}
	s.crypto = cb
	return nil
}

// Crypto returns the crypto backend
func (s *Store) Crypto() backend.Crypto {
	return s.crypto
}

// ImportMissingPublicKeys will try to import any missing public keys from the
// .public-keys folder in the password store
func (s *Store) ImportMissingPublicKeys(ctx context.Context) error {
	// do not import any keys for age, where public key == key id
	if _, ok := s.crypto.(*age.Age); ok {
		debug.Log("not importing public keys for age")
		return nil
	}
	rs, err := s.GetRecipients(ctx, "")
	if err != nil {
		return errors.Wrapf(err, "failed to get recipients")
	}
	for _, r := range rs {
		debug.Log("Checking recipients %s ...", r)
		// check if this recipient is missing
		// we could list all keys outside the loop and just do the lookup here
		// but this way we ensure to use the exact same lookup logic as
		// gpg does on encryption
		kl, err := s.crypto.FindRecipients(ctx, r)
		if err != nil {
			out.Error(ctx, "[%s] Failed to get public key for %s: %s", s.alias, r, err)
		}
		if len(kl) > 0 {
			debug.Log("[%s] Keyring contains %d public keys for %s", s.alias, len(kl), r)
			continue
		}

		// get info about this public key
		names, err := s.decodePublicKey(ctx, r)
		if err != nil {
			out.Error(ctx, "[%s] Failed to decode public key %s: %s", s.alias, r, err)
			continue
		}

		// we need to ask the user before importing
		// any key material into his keyring!
		if imf := ctxutil.GetImportFunc(ctx); imf != nil {
			if !imf(ctx, r, names) {
				continue
			}
		}

		debug.Log("[%s] Public Key %s not found in keyring, importing", s.alias, r)

		// try to load this recipient
		if err := s.importPublicKey(ctx, r); err != nil {
			out.Error(ctx, "[%s] Failed to import public key for %s: %s", s.alias, r, err)
			continue
		}
		out.Green(ctx, "[%s] Imported public key for %s into Keyring", s.alias, r)
	}
	return nil
}

func (s *Store) decodePublicKey(ctx context.Context, r string) ([]string, error) {
	for _, kd := range []string{keyDir, oldKeyDir} {
		filename := filepath.Join(kd, r)
		if !s.storage.Exists(ctx, filename) {
			debug.Log("Public Key %s not found at %s", r, filename)
			continue
		}
		buf, err := s.storage.Get(ctx, filename)
		if err != nil {
			return nil, errors.Errorf("Unable to read Public Key %s %s: %s", r, filename, err)
		}
		return s.crypto.ReadNamesFromKey(ctx, buf)
	}
	return nil, errors.Errorf("Public Key %s not found", r)
}

// export an ASCII armored public key
func (s *Store) exportPublicKey(ctx context.Context, r string) (string, error) {
	filename := filepath.Join(keyDir, r)

	// do not overwrite existing keys
	if s.storage.Exists(ctx, filename) {
		return "", nil
	}

	pk, err := s.crypto.ExportPublicKey(ctx, r)
	if err != nil {
		return "", errors.Wrapf(err, "failed to export public key")
	}

	// ECC keys are at least 700 byte, RSA should be a lot bigger
	if len(pk) < 32 {
		return "", errors.New("exported key too small")
	}

	if err := s.storage.Set(ctx, filename, pk); err != nil {
		return "", errors.Wrapf(err, "failed to write exported public key to store")
	}

	return filename, nil
}

// import an public key into the default keyring
func (s *Store) importPublicKey(ctx context.Context, r string) error {
	for _, kd := range []string{keyDir, oldKeyDir} {
		filename := filepath.Join(kd, r)
		if !s.storage.Exists(ctx, filename) {
			debug.Log("Public Key %s not found at %s", r, filename)
			continue
		}
		pk, err := s.storage.Get(ctx, filename)
		if err != nil {
			return err
		}
		return s.crypto.ImportPublicKey(ctx, pk)
	}
	return fmt.Errorf("public key not found in store")
}
