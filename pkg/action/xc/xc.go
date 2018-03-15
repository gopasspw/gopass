package xc

import (
	"context"
	"io/ioutil"

	"github.com/justwatchcom/gopass/pkg/action"
	"github.com/justwatchcom/gopass/pkg/agent/client"
	"github.com/justwatchcom/gopass/pkg/backend/crypto/xc"
	"github.com/justwatchcom/gopass/pkg/config"
	"github.com/justwatchcom/gopass/pkg/fsutil"
	"github.com/justwatchcom/gopass/pkg/out"
	"github.com/justwatchcom/gopass/pkg/termio"
	"github.com/urfave/cli"
)

// ListPrivateKeys list the XC private keys
func ListPrivateKeys(ctx context.Context, c *cli.Context) error {
	cfgdir := config.Directory()
	crypto, err := xc.New(cfgdir, client.New(cfgdir))
	if err != nil {
		return action.ExitError(ctx, action.ExitUnknown, err, "failed to init XC")
	}

	kl, err := crypto.ListPrivateKeyIDs(ctx)
	if err != nil {
		return action.ExitError(ctx, action.ExitUnknown, err, "failed to list private keys")
	}

	out.Print(ctx, "XC Private Keys:")
	for _, key := range kl {
		out.Print(ctx, "%s - %s", key, crypto.FormatKey(ctx, key))
	}

	return nil
}

// ListPublicKeys lists the XC public keys
func ListPublicKeys(ctx context.Context, c *cli.Context) error {
	cfgdir := config.Directory()
	crypto, err := xc.New(cfgdir, client.New(cfgdir))
	if err != nil {
		return action.ExitError(ctx, action.ExitUnknown, err, "failed to init XC")
	}

	kl, err := crypto.ListPublicKeyIDs(ctx)
	if err != nil {
		return action.ExitError(ctx, action.ExitUnknown, err, "failed to list public keys")
	}

	out.Print(ctx, "XC Public Keys:")
	for _, key := range kl {
		out.Print(ctx, "%s - %s", key, crypto.FormatKey(ctx, key))
	}

	return nil
}

// GenerateKeypair generates a new XC keypair
func GenerateKeypair(ctx context.Context, c *cli.Context) error {
	name := c.String("name")
	email := c.String("email")
	pw := c.String("passphrase")

	if name == "" {
		var err error
		name, err = termio.AskForString(ctx, "What is your full name?", "")
		if err != nil || name == "" {
			return action.ExitError(ctx, action.ExitNoName, err, "please provide a name")
		}
	}
	if email == "" {
		var err error
		email, err = termio.AskForString(ctx, "What is your email?", "")
		if err != nil || name == "" {
			return action.ExitError(ctx, action.ExitNoName, err, "please provide a email")
		}
	}
	if pw == "" {
		var err error
		pw, err = termio.AskForPassword(ctx, name)
		if err != nil {
			return action.ExitError(ctx, action.ExitIO, err, "failed to ask for password: %s", err)
		}
	}

	cfgdir := config.Directory()
	crypto, err := xc.New(cfgdir, client.New(cfgdir))
	if err != nil {
		return action.ExitError(ctx, action.ExitUnknown, err, "failed to init XC")
	}

	return crypto.CreatePrivateKeyBatch(ctx, name, email, pw)
}

// ExportPublicKey exports an XC key
func ExportPublicKey(ctx context.Context, c *cli.Context) error {
	id := c.String("id")
	file := c.String("file")

	if id == "" {
		return action.ExitError(ctx, action.ExitUsage, nil, "need id")
	}
	if file == "" {
		return action.ExitError(ctx, action.ExitUsage, nil, "need file")
	}

	cfgdir := config.Directory()
	crypto, err := xc.New(cfgdir, client.New(cfgdir))
	if err != nil {
		return action.ExitError(ctx, action.ExitUnknown, err, "failed to init XC")
	}

	if fsutil.IsFile(file) {
		return action.ExitError(ctx, action.ExitUnknown, nil, "output file already exists")
	}

	pk, err := crypto.ExportPublicKey(ctx, id)
	if err != nil {
		return action.ExitError(ctx, action.ExitUnknown, err, "failed to export key: %s", err)
	}

	if err := ioutil.WriteFile(file, pk, 0600); err != nil {
		return action.ExitError(ctx, action.ExitIO, err, "failed to write file")
	}
	return nil
}

// ImportPublicKey imports an XC key
func ImportPublicKey(ctx context.Context, c *cli.Context) error {
	file := c.String("file")

	if file == "" {
		return action.ExitError(ctx, action.ExitUsage, nil, "need file")
	}

	cfgdir := config.Directory()
	crypto, err := xc.New(cfgdir, client.New(cfgdir))
	if err != nil {
		return action.ExitError(ctx, action.ExitUnknown, err, "failed to init XC")
	}

	if !fsutil.IsFile(file) {
		return action.ExitError(ctx, action.ExitNotFound, nil, "input file not found")
	}

	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return action.ExitError(ctx, action.ExitIO, err, "failed to read file")
	}
	return crypto.ImportPublicKey(ctx, buf)
}

// RemoveKey removes a key from the keyring
func RemoveKey(ctx context.Context, c *cli.Context) error {
	id := c.String("id")

	if id == "" {
		return action.ExitError(ctx, action.ExitUsage, nil, "need id")
	}

	cfgdir := config.Directory()
	crypto, err := xc.New(cfgdir, client.New(cfgdir))
	if err != nil {
		return action.ExitError(ctx, action.ExitUnknown, err, "failed to init XC")
	}

	return crypto.RemoveKey(id)
}

// ExportPrivateKey exports an XC key
func ExportPrivateKey(ctx context.Context, c *cli.Context) error {
	id := c.String("id")
	file := c.String("file")

	if id == "" {
		return action.ExitError(ctx, action.ExitUsage, nil, "need id")
	}
	if file == "" {
		return action.ExitError(ctx, action.ExitUsage, nil, "need file")
	}

	cfgdir := config.Directory()
	crypto, err := xc.New(cfgdir, client.New(cfgdir))
	if err != nil {
		return action.ExitError(ctx, action.ExitUnknown, err, "failed to init XC")
	}

	if fsutil.IsFile(file) {
		return action.ExitError(ctx, action.ExitUnknown, nil, "output file already exists")
	}

	pk, err := crypto.ExportPrivateKey(ctx, id)
	if err != nil {
		return action.ExitError(ctx, action.ExitUnknown, err, "failed to export key: %s", err)
	}

	if err := ioutil.WriteFile(file, pk, 0600); err != nil {
		return action.ExitError(ctx, action.ExitIO, err, "failed to write file")
	}
	return nil
}

// ImportPrivateKey imports an XC key
func ImportPrivateKey(ctx context.Context, c *cli.Context) error {
	file := c.String("file")

	if file == "" {
		return action.ExitError(ctx, action.ExitUsage, nil, "need file")
	}

	cfgdir := config.Directory()
	crypto, err := xc.New(cfgdir, client.New(cfgdir))
	if err != nil {
		return action.ExitError(ctx, action.ExitUnknown, err, "failed to init XC")
	}

	if !fsutil.IsFile(file) {
		return action.ExitError(ctx, action.ExitNotFound, nil, "input file not found")
	}

	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return action.ExitError(ctx, action.ExitIO, err, "failed to read file")
	}
	return crypto.ImportPrivateKey(ctx, buf)
}
