package backend

import (
	"context"

	"github.com/blang/semver"
)

// CryptoBackend is a cryptographic backend
type CryptoBackend int

const (
	// GPGMock is a no-op crypto backend
	GPGMock CryptoBackend = iota
	// GPGCLI is a gpg-cli based crypto backend
	GPGCLI
	// XC is an experimental crypto backend
	XC
	// Trezor is a USB hardware secure processor crypto backend
	TREZOR
)

// Keyring is a public/private key manager
type Keyring interface {
	ImportPublicKey(ctx context.Context, key []byte) error
	ExportPublicKey(ctx context.Context, id string) ([]byte, error)

	ListPublicKeyIDs(ctx context.Context) ([]string, error)
	ListPrivateKeyIDs(ctx context.Context) ([]string, error)

	FindPublicKeys(ctx context.Context, needles ...string) ([]string, error)
	FindPrivateKeys(ctx context.Context, needles ...string) ([]string, error)

	FormatKey(ctx context.Context, id string) string
	NameFromKey(ctx context.Context, id string) string
	EmailFromKey(ctx context.Context, id string) string
	ReadNamesFromKey(ctx context.Context, buf []byte) ([]string, error)

	CreatePrivateKeyBatch(ctx context.Context, name, email, passphrase string) error
	CreatePrivateKey(ctx context.Context) error
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
