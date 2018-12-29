// +build !xc

package xc

import (
	"context"

	"github.com/urfave/cli"
)

// ListPrivateKeys list the XC private keys
func ListPrivateKeys(ctx context.Context, c *cli.Context) error {
	return nil
}

// ListPublicKeys lists the XC public keys
func ListPublicKeys(ctx context.Context, c *cli.Context) error {
	return nil
}

// GenerateKeypair generates a new XC keypair
func GenerateKeypair(ctx context.Context, c *cli.Context) error {
	return nil
}

// ExportPublicKey exports an XC key
func ExportPublicKey(ctx context.Context, c *cli.Context) error {
	return nil
}

// ImportPublicKey imports an XC key
func ImportPublicKey(ctx context.Context, c *cli.Context) error {
	return nil
}

// RemoveKey removes a key from the keyring
func RemoveKey(ctx context.Context, c *cli.Context) error {
	return nil
}

// ExportPrivateKey exports an XC key
func ExportPrivateKey(ctx context.Context, c *cli.Context) error {
	return nil
}

// ImportPrivateKey imports an XC key
func ImportPrivateKey(ctx context.Context, c *cli.Context) error {
	return nil
}

// EncryptFile encrypts a single file
func EncryptFile(ctx context.Context, c *cli.Context) error {
	return nil
}

// DecryptFile decrypts a single file
func DecryptFile(ctx context.Context, c *cli.Context) error {
	return nil
}

// EncryptFileStream encrypts a single file
func EncryptFileStream(ctx context.Context, c *cli.Context) error {
	return nil
}

// DecryptFileStream decrypts a single file
func DecryptFileStream(ctx context.Context, c *cli.Context) error {
	return nil
}
