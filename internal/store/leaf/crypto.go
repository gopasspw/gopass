package leaf

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/pkg/debug"
)

func (s *Store) initCryptoBackend(ctx context.Context) error {
	cb, err := backend.DetectCrypto(ctx, s.storage)
	if err != nil {
		return err
	}
	s.crypto = cb

	return nil
}

// Crypto returns the crypto backend.
func (s *Store) Crypto() backend.Crypto {
	return s.crypto
}

// recipientCheck checks if a recipient is already present in the keyring and up-to-date.
// It returns true if the recipient is fine and the import can be skipped.
//
// Stage 4 (GH-1430): When FindRecipients returns empty (key expired / unusable
// in keyring) but the fingerprint matches a key in the store's .public-keys/,
// this returns false so ImportMissingPublicKeys can re-import the fresh key.
func (s *Store) recipientCheck(ctx context.Context, r string) bool {
	// check if this recipient is missing
	// we could list all keys outside the loop and just do the lookup here
	// but this way we ensure to use the exact same lookup logic as
	// gpg does on encryption
	kl, err := s.crypto.FindRecipients(ctx, r)
	if err != nil {
		// this is expected if we don't have the key
		debug.Log("Failed to get public key for %s: %s", r, err)
	}

	if len(kl) > 0 { //nolint:nestif
		debug.Log("Keyring contains %d public keys for %s", len(kl), r)
		if !IsPubkeyUpdate(ctx) {
			return true
		}
		ex, ok := s.crypto.(keyExporter)
		if !ok {
			return true
		}
		pk, err := ex.ExportPublicKey(ctx, r)
		if err != nil {
			return true
		}
		pk2, err2 := s.getPublicKey(ctx, r)
		if err2 != nil {
			return true
		}
		if bytes.Equal(pk, pk2) {
			return true
		}
	} else {
		// Key not found in keyring (may be expired/unusable). Try to
		// look it up by fingerprint from the .public-keys/ copy.
		pk, err := s.getPublicKey(ctx, r)
		if err != nil {
			debug.Log("failed to get public key for %s: %s", r, err)

			return true
		}
		fp, err := s.crypto.GetFingerprint(ctx, pk)
		if err != nil {
			debug.Log("failed to get fingerprint for %s: %s", r, err)

			return true
		}
		kl, err = s.crypto.FindRecipients(ctx, fp)
		if err != nil {
			debug.Log("failed to find recipients for %s: %s", fp, err)
		}

		if len(kl) > 0 {
			// Stage 4 (GH-1430): key found by fingerprint but
			// FindRecipients on the original ID returned empty
			// (e.g. expired). Re-import so the keyring gets the
			// fresh copy from .public-keys/.
			debug.Log("key %s with fingerprint %s found in keyring but original ID not usable (expired?); will re-import", r, fp)

			return false
		}
	}

	return false
}

// ImportMissingPublicKeys will try to import any missing public keys from the
// .public-keys folder in the password store.
func (s *Store) ImportMissingPublicKeys(ctx context.Context, newrs ...string) error {
	// only import public keys for backends that manage a separate keyring
	if !s.crypto.NeedsPublicKeyImport() {
		debug.Log("not importing public keys for %s (not needed by this backend)", s.crypto.Name())

		return nil
	}

	rs, err := s.GetRecipients(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to get recipients: %w", err)
	}

	ids := append(rs.IDs(), newrs...)
	for _, r := range ids {
		debug.Log("Checking recipients %s ...", r)
		if s.recipientCheck(ctx, r) {
			continue
		}

		// get info about this public key
		names, err := s.decodePublicKey(ctx, r)
		if err != nil {
			out.Errorf(ctx, "Failed to decode public key %s: %s", r, err)

			continue
		}

		// we need to ask the user before importing
		// any key material into his keyring!
		if s.importCallback != nil && !config.Bool(ctx, "core.autoimport") {
			if !s.importCallback(ctx, r, names) {
				continue
			}
		}

		debug.Log("Public Key %s not found in keyring, importing", r)

		// try to load this recipient
		if err := s.importPublicKey(ctx, r); err != nil {
			out.Errorf(ctx, "Failed to import public key for %s: %s", r, err)

			continue
		}

		// Stage 4 (GH-1430): distinguish fresh import from update.
		// If the key was already in the keyring (FindRecipients returned
		// empty but fingerprint matched — expired), say "updated".
		kl, _ := s.crypto.FindRecipients(ctx, r)
		if len(kl) == 0 {
			out.Printf(ctx, "Imported public key for %s into Keyring", r)
		} else {
			out.Printf(ctx, "Updated public key for %s in Keyring", r)
		}
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
			return nil, fmt.Errorf("unable to read Public Key %q %q: %w", r, filename, err)
		}

		return s.crypto.ReadNamesFromKey(ctx, buf)
	}

	return nil, fmt.Errorf("public key %q not found", r)
}

// export an ASCII armored public key.
func (s *Store) exportPublicKey(ctx context.Context, exp keyExporter, r string) (string, error) {
	filename := filepath.Join(keyDir, r)

	// do not overwrite existing keys, unless forced
	if !IsPubkeyUpdate(ctx) && s.storage.Exists(ctx, filename) {
		debug.Log("leaving existing key for %s at %s alone", filename)

		return "", nil
	}

	pk, err := exp.ExportPublicKey(ctx, r)
	if err != nil {
		return "", fmt.Errorf("failed to export public key: %w", err)
	}

	// ECC keys are at least 700 byte, RSA should be a lot bigger
	if len(pk) < 32 {
		return "", fmt.Errorf("exported key too small")
	}

	if err := s.storage.Set(ctx, filename, pk); err != nil {
		if !errors.Is(err, store.ErrMeaninglessWrite) {
			return "", fmt.Errorf("failed to write exported public key to store: %w", err)
		}
		debug.Log("No need to write exported public key %s: already stored", r)
	}

	debug.Log("exported public keys for %s to %s", r, filename)

	return filename, nil
}

type keyImporter interface {
	ImportPublicKey(ctx context.Context, key []byte) error
}
type keyExporter interface {
	ExportPublicKey(ctx context.Context, id string) ([]byte, error)
}

func (s *Store) getPublicKey(ctx context.Context, r string) ([]byte, error) {
	for _, kd := range []string{keyDir, oldKeyDir} {
		filename := filepath.Join(kd, r)
		if !s.storage.Exists(ctx, filename) {
			debug.Log("Public Key %s not found at %s", r, filename)

			continue
		}
		pk, err := s.storage.Get(ctx, filename)

		return pk, err
	}

	return nil, fmt.Errorf("public key not found in store")
}

// import an public key into the default keyring.
func (s *Store) importPublicKey(ctx context.Context, r string) error {
	im, ok := s.crypto.(keyImporter)
	if !ok {
		debug.Log("importing public keys not supported by %T", s.crypto)

		return nil
	}

	pk, err := s.getPublicKey(ctx, r)
	if err != nil {
		return err
	}

	return im.ImportPublicKey(ctx, pk)
}

type locker interface {
	Lock()
}

// Lock clears the credential caches of all supported backends.
func (s *Store) Lock() error {
	f, ok := s.crypto.(locker)
	if !ok {
		debug.Log("locking not supported by %T in %q", s.crypto, s.alias)
	}

	if f == nil {
		debug.Log("backend %q invalid", s.alias)

		return nil
	}

	f.Lock()
	debug.Log("locked backend %T for %q", s.crypto, s.alias)

	return nil
}
