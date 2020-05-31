// +build xc

package xc

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/gopasspw/gopass/internal/action"
	"github.com/gopasspw/gopass/internal/backend/crypto/xc"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/termio"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/fsutil"
	"github.com/urfave/cli/v2"
)

var crypto *xc.XC

func initCrypto() error {
	if crypto != nil {
		return nil
	}

	cfgdir := config.Directory()
	crypto = xc.New(cfgdir, nil)
	return nil
}

// ListPrivateKeys list the XC private keys
func ListPrivateKeys(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	if err := initCrypto(); err != nil {
		return action.ExitError(action.ExitUnknown, err, "failed to init XC")
	}

	kl, err := crypto.ListIdentities(ctx)
	if err != nil {
		return action.ExitError(action.ExitUnknown, err, "failed to list private keys")
	}

	out.Print(ctx, "XC Private Keys:")
	for _, key := range kl {
		out.Print(ctx, "%s - %s", key, crypto.FormatKey(ctx, key, ""))
	}

	return nil
}

// ListPublicKeys lists the XC public keys
func ListPublicKeys(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	if err := initCrypto(); err != nil {
		return action.ExitError(action.ExitUnknown, err, "failed to init XC")
	}

	kl, err := crypto.ListRecipients(ctx)
	if err != nil {
		return action.ExitError(action.ExitUnknown, err, "failed to list public keys")
	}

	out.Print(ctx, "XC Public Keys:")
	for _, key := range kl {
		out.Print(ctx, "%s - %s", key, crypto.FormatKey(ctx, key, ""))
	}

	return nil
}

// GenerateKeypair generates a new XC keypair
func GenerateKeypair(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	if err := initCrypto(); err != nil {
		return action.ExitError(action.ExitUnknown, err, "failed to init XC")
	}

	name := c.String("name")
	email := c.String("email")
	pw := c.String("passphrase")

	if name == "" {
		var err error
		name, err = termio.AskForString(ctx, "What is your full name?", "")
		if err != nil || name == "" {
			return action.ExitError(action.ExitNoName, err, "please provide a name")
		}
	}
	if email == "" {
		var err error
		email, err = termio.AskForString(ctx, "What is your email?", "")
		if err != nil || email == "" {
			return action.ExitError(action.ExitNoName, err, "please provide an email")
		}
	}
	if pw == "" {
		var err error
		pw, err = termio.AskForPassword(ctx, "")
		if err != nil || pw == "" {
			return action.ExitError(action.ExitIO, err, "failed to ask for password: %s", err)
		}
	}

	if err := crypto.GenerateIdentity(ctx, name, email, pw); err != nil {
		return action.ExitError(action.ExitUnknown, err, "failed to create private key: %s", err)
	}
	return nil
}

// ExportPublicKey exports an XC key
func ExportPublicKey(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	if err := initCrypto(); err != nil {
		return action.ExitError(action.ExitUnknown, err, "failed to init XC")
	}

	id := c.String("id")
	file := c.String("file")

	if id == "" {
		return action.ExitError(action.ExitUsage, nil, "need id")
	}
	if file == "" {
		return action.ExitError(action.ExitUsage, nil, "need file")
	}

	if fsutil.IsFile(file) {
		return action.ExitError(action.ExitUnknown, nil, "output file already exists")
	}

	pk, err := crypto.ExportPublicKey(ctx, id)
	if err != nil {
		return action.ExitError(action.ExitUnknown, err, "failed to export key: %s", err)
	}

	if err := ioutil.WriteFile(file, pk, 0600); err != nil {
		return action.ExitError(action.ExitIO, err, "failed to write file")
	}
	return nil
}

// ImportPublicKey imports an XC key
func ImportPublicKey(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	if err := initCrypto(); err != nil {
		return action.ExitError(action.ExitUnknown, err, "failed to init XC")
	}

	file := c.String("file")

	if file == "" {
		return action.ExitError(action.ExitUsage, nil, "need file")
	}

	if !fsutil.IsFile(file) {
		return action.ExitError(action.ExitNotFound, nil, "input file not found")
	}

	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return action.ExitError(action.ExitIO, err, "failed to read file")
	}
	return crypto.ImportPublicKey(ctx, buf)
}

// RemoveKey removes a key from the keyring
func RemoveKey(c *cli.Context) error {
	if err := initCrypto(); err != nil {
		return action.ExitError(action.ExitUnknown, err, "failed to init XC")
	}

	id := c.String("id")

	if id == "" {
		return action.ExitError(action.ExitUsage, nil, "need id")
	}

	return crypto.RemoveKey(id)
}

// ExportPrivateKey exports an XC key
func ExportPrivateKey(c *cli.Context) error {
	if err := initCrypto(); err != nil {
		return action.ExitError(action.ExitUnknown, err, "failed to init XC")
	}

	id := c.String("id")
	file := c.String("file")

	if id == "" {
		return action.ExitError(action.ExitUsage, nil, "need id")
	}
	if file == "" {
		return action.ExitError(action.ExitUsage, nil, "need file")
	}

	if fsutil.IsFile(file) {
		return action.ExitError(action.ExitUnknown, nil, "output file already exists")
	}

	pk, err := crypto.ExportPrivateKey(id)
	if err != nil {
		return action.ExitError(action.ExitUnknown, err, "failed to export key: %s", err)
	}

	if err := ioutil.WriteFile(file, pk, 0600); err != nil {
		return action.ExitError(action.ExitIO, err, "failed to write file")
	}
	return nil
}

// ImportPrivateKey imports an XC key
func ImportPrivateKey(c *cli.Context) error {
	if err := initCrypto(); err != nil {
		return action.ExitError(action.ExitUnknown, err, "failed to init XC")
	}

	file := c.String("file")

	if file == "" {
		return action.ExitError(action.ExitUsage, nil, "need file")
	}

	if !fsutil.IsFile(file) {
		return action.ExitError(action.ExitNotFound, nil, "input file not found")
	}

	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return action.ExitError(action.ExitIO, err, "failed to read file")
	}
	return crypto.ImportPrivateKey(buf)
}

// EncryptFile encrypts a single file
func EncryptFile(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	if err := initCrypto(); err != nil {
		return action.ExitError(action.ExitUnknown, err, "failed to init XC")
	}

	if c.Bool("stream") {
		return EncryptFileStream(c)
	}

	inFile := c.String("file")
	if inFile == "" {
		return action.ExitError(action.ExitUsage, nil, "need file")
	}

	recipients := c.StringSlice("recipients")
	outFile := inFile + ".xc"

	if !fsutil.IsFile(inFile) {
		return action.ExitError(action.ExitNotFound, nil, "input file not found")
	}
	if fsutil.IsFile(outFile) {
		return action.ExitError(action.ExitIO, nil, "output file already exists")
	}

	plaintext, err := ioutil.ReadFile(inFile)
	if err != nil {
		return action.ExitError(action.ExitIO, err, "failed to read file")
	}
	ciphertext, err := crypto.Encrypt(ctx, plaintext, recipients)
	if err != nil {
		return action.ExitError(action.ExitUnknown, err, "failed to encrypt file: %s", err)
	}
	if err := ioutil.WriteFile(outFile, ciphertext, 0600); err != nil {
		return action.ExitError(action.ExitUnknown, err, "failed to write ciphertext")
	}
	return nil
}

// DecryptFile decrypts a single file
func DecryptFile(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	if err := initCrypto(); err != nil {
		return action.ExitError(action.ExitUnknown, err, "failed to init XC")
	}

	if c.Bool("stream") {
		return DecryptFileStream(c)
	}

	inFile := c.String("file")
	if inFile == "" {
		return action.ExitError(action.ExitUsage, nil, "need file")
	}
	if !strings.HasSuffix(inFile, ".xc") {
		return action.ExitError(action.ExitUsage, nil, "unknown extension. expecting .xc")
	}
	outFile := strings.TrimSuffix(inFile, ".xc")

	if !fsutil.IsFile(inFile) {
		return action.ExitError(action.ExitNotFound, nil, "input file not found")
	}
	if fsutil.IsFile(outFile) {
		return action.ExitError(action.ExitIO, nil, "output file already exists")
	}

	ciphertext, err := ioutil.ReadFile(inFile)
	if err != nil {
		return action.ExitError(action.ExitIO, err, "failed to read file")
	}

	plaintext, err := crypto.Decrypt(ctx, ciphertext)
	if err != nil {
		return action.ExitError(action.ExitUnknown, err, "failed to decrypt file: %s", err)
	}

	if err := ioutil.WriteFile(outFile, plaintext, 0600); err != nil {
		return action.ExitError(action.ExitUnknown, err, "failed to write plaintext")
	}
	return nil
}

// EncryptFileStream encrypts a single file
func EncryptFileStream(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	if err := initCrypto(); err != nil {
		return action.ExitError(action.ExitUnknown, err, "failed to init XC")
	}

	inFile := c.String("file")
	if inFile == "" {
		return action.ExitError(action.ExitUsage, nil, "need file")
	}

	recipients := c.StringSlice("recipients")

	outFile := inFile + ".xc"

	if !fsutil.IsFile(inFile) {
		return action.ExitError(action.ExitNotFound, nil, "input file not found")
	}
	if fsutil.IsFile(outFile) {
		return action.ExitError(action.ExitIO, nil, "output file already exists")
	}

	plaintext, err := os.Open(inFile)
	if err != nil {
		return action.ExitError(action.ExitIO, err, "failed to open file")
	}
	defer func() { _ = plaintext.Close() }()

	ciphertext, err := os.OpenFile(outFile, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return action.ExitError(action.ExitIO, err, "failed to open file")
	}
	defer func() { _ = ciphertext.Close() }()

	if err := crypto.EncryptStream(ctx, plaintext, recipients, ciphertext); err != nil {
		_ = os.Remove(outFile)
		return action.ExitError(action.ExitUnknown, err, "failed to encrypt file: %s", err)
	}
	return nil
}

// DecryptFileStream decrypts a single file
func DecryptFileStream(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	if err := initCrypto(); err != nil {
		return action.ExitError(action.ExitUnknown, err, "failed to init XC")
	}

	inFile := c.String("file")
	if inFile == "" {
		return action.ExitError(action.ExitUsage, nil, "need file")
	}
	if !strings.HasSuffix(inFile, ".xc") {
		return action.ExitError(action.ExitUsage, nil, "unknown extension. expecting .xc")
	}
	outFile := strings.TrimSuffix(inFile, ".xc")

	if !fsutil.IsFile(inFile) {
		return action.ExitError(action.ExitNotFound, nil, "input file not found")
	}
	if fsutil.IsFile(outFile) {
		return action.ExitError(action.ExitIO, nil, "output file already exists")
	}

	ciphertext, err := os.Open(inFile)
	if err != nil {
		return action.ExitError(action.ExitIO, err, "failed to read file")
	}
	defer func() { _ = ciphertext.Close() }()

	plaintext, err := os.OpenFile(outFile, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return action.ExitError(action.ExitIO, err, "failed to read file")
	}
	defer func() { _ = plaintext.Close() }()

	if err := crypto.DecryptStream(ctx, ciphertext, plaintext); err != nil {
		_ = os.Remove(outFile)
		return action.ExitError(action.ExitUnknown, err, "failed to decrypt file: %s", err)
	}
	return nil
}
