package backend

import (
	"context"
	"fmt"

	"github.com/blang/semver/v4"
	"github.com/gopasspw/gopass/pkg/debug"
)

// CryptoBackend is a cryptographic backend.
type CryptoBackend int

const (
	// Plain is a no-op crypto backend.
	Plain CryptoBackend = iota
	// GPGCLI is a gpg-cli based crypto backend.
	GPGCLI
	// Age - age-encryption.org.
	Age
)

func (c CryptoBackend) String() string {
	if be, err := CryptoRegistry.BackendName(c); err == nil {
		return be
	}

	return ""
}

// Keyring is a public/private key manager.
type Keyring interface {
	ListRecipients(ctx context.Context) ([]string, error)
	ListIdentities(ctx context.Context) ([]string, error)

	FindRecipients(ctx context.Context, needles ...string) ([]string, error)
	FindIdentities(ctx context.Context, needles ...string) ([]string, error)

	Fingerprint(ctx context.Context, id string) string
	FormatKey(ctx context.Context, id, tpl string) string
	ReadNamesFromKey(ctx context.Context, buf []byte) ([]string, error)
	GetFingerprint(ctx context.Context, key []byte) (string, error)

	GenerateIdentity(ctx context.Context, name, email, passphrase string) (string, error)
}

// Crypto is a crypto backend.
type Crypto interface {
	Keyring

	Encrypt(ctx context.Context, plaintext []byte, recipients []string) ([]byte, error)
	Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error)
	RecipientIDs(ctx context.Context, ciphertext []byte) ([]string, error)

	Name() string
	Version(context.Context) semver.Version
	Initialized(ctx context.Context) error
	Ext() string    // filename extension.
	IDFile() string // recipient IDs.
	Concurrency() int
	// NeedsPublicKeyImport reports whether this backend manages public keys in a
	// separate keyring that must be populated via import. GPG-style backends
	// return true; backends such as age, where the public key IS the recipient
	// identifier, return false.
	NeedsPublicKeyImport() bool
}

// NewCrypto instantiates a new crypto backend.
func NewCrypto(ctx context.Context, id CryptoBackend) (Crypto, error) {
	if be, err := CryptoRegistry.Get(id); err == nil {
		return be.New(ctx)
	}

	return nil, fmt.Errorf("unknown backend %d: %w", id, ErrNotFound)
}

// DetectCrypto tries to detect the crypto backend used.
// If a backend is explicitly requested via the context and cannot be found,
// an error is returned. If no backend is requested and none can be detected
// from the storage (i.e. the store is uninitialized), nil, nil is returned
// so that callers can distinguish an uninitialized store from a hard error.
func DetectCrypto(ctx context.Context, storage Storage) (Crypto, error) {
	if HasCryptoBackend(ctx) {
		be, err := CryptoRegistry.Get(GetCryptoBackend(ctx))
		if err == nil {
			return be.New(ctx)
		}

		// An explicit backend was requested but is not registered.
		return nil, fmt.Errorf("no crypto backend found for %s: %w", storage, ErrNotFound)
	}

	for _, be := range CryptoRegistry.Prioritized() {
		debug.Log("Trying %s for %s", be, storage)
		if err := be.Handles(ctx, storage); err != nil {
			debug.Log("failed to use crypto %s for %s", be, storage)

			continue
		}
		debug.Log("Using %s for %s", be, storage)

		return be.New(ctx)
	}

	// No backend could be auto-detected. The store is likely uninitialized.
	debug.Log("No valid crypto provider found for %s", storage)

	return nil, nil //nolint:nilnil
}
