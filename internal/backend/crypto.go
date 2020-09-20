package backend

import (
	"context"
	"sort"

	"github.com/blang/semver"
	"github.com/gopasspw/gopass/internal/debug"
	"github.com/pkg/errors"
)

// CryptoBackend is a cryptographic backend
type CryptoBackend int

const (
	// Plain is a no-op crypto backend
	Plain CryptoBackend = iota
	// GPGCLI is a gpg-cli based crypto backend
	GPGCLI
	// Age - age-encryption.org
	Age
)

func (c CryptoBackend) String() string {
	return CryptoNameFromBackend(c)
}

// Keyring is a public/private key manager
type Keyring interface {
	ImportPublicKey(ctx context.Context, key []byte) error
	ExportPublicKey(ctx context.Context, id string) ([]byte, error)

	ListRecipients(ctx context.Context) ([]string, error)
	ListIdentities(ctx context.Context) ([]string, error)

	FindRecipients(ctx context.Context, needles ...string) ([]string, error)
	FindIdentities(ctx context.Context, needles ...string) ([]string, error)

	Fingerprint(ctx context.Context, id string) string
	FormatKey(ctx context.Context, id, tpl string) string
	ReadNamesFromKey(ctx context.Context, buf []byte) ([]string, error)

	GenerateIdentity(ctx context.Context, name, email, passphrase string) error
}

// Crypto is a crypto backend
type Crypto interface {
	Keyring

	Encrypt(ctx context.Context, plaintext []byte, recipients []string) ([]byte, error)
	Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error)
	RecipientIDs(ctx context.Context, ciphertext []byte) ([]string, error)

	Name() string
	Version(context.Context) semver.Version
	Initialized(ctx context.Context) error
	Ext() string    // filename extension
	IDFile() string // recipient IDs
}

// RegisterCrypto registers a new crypto backend with the backend registry.
func RegisterCrypto(id CryptoBackend, name string, loader CryptoLoader) {
	cryptoRegistry[id] = loader
	cryptoNameToBackendMap[name] = id
	cryptoBackendToNameMap[id] = name
}

// NewCrypto instantiates a new crypto backend.
func NewCrypto(ctx context.Context, id CryptoBackend) (Crypto, error) {
	if be, found := cryptoRegistry[id]; found {
		return be.New(ctx)
	}
	return nil, errors.Wrapf(ErrNotFound, "unknown backend: %d", id)
}

// DetectCrypto tries to detect the crypto backend used
func DetectCrypto(ctx context.Context, storage Storage) (Crypto, error) {
	if HasCryptoBackend(ctx) {
		if be, found := cryptoRegistry[GetCryptoBackend(ctx)]; found {
			return be.New(ctx)
		}
	}

	bes := make([]CryptoBackend, 0, len(cryptoRegistry))
	for id := range cryptoRegistry {
		bes = append(bes, id)
	}
	sort.Slice(bes, func(i, j int) bool {
		return cryptoRegistry[bes[i]].Priority() < cryptoRegistry[bes[j]].Priority()
	})
	for _, id := range bes {
		be := cryptoRegistry[id]
		debug.Log("Trying %s for %s", be, storage)
		if err := be.Handles(storage); err != nil {
			debug.Log("failed to use crypto %s for %s", id, storage)
			continue
		}
		debug.Log("Using %s for %s", be, storage)
		return be.New(ctx)
	}
	debug.Log("No valid crypto provider found for %s", storage)
	return nil, nil
}
