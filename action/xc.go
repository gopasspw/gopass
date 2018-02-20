package action

import (
	"context"
	"io/ioutil"

	"github.com/justwatchcom/gopass/backend/crypto/xc"
	"github.com/justwatchcom/gopass/config"
	"github.com/justwatchcom/gopass/utils/agent/client"
	"github.com/justwatchcom/gopass/utils/fsutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/justwatchcom/gopass/utils/termio"
	"github.com/urfave/cli"
)

// XCListPrivateKeys list the XC private keys
func (s *Action) XCListPrivateKeys(ctx context.Context, c *cli.Context) error {
	cfgdir := config.Directory()
	crypto, err := xc.New(cfgdir, client.New(cfgdir))
	if err != nil {
		return exitError(ctx, ExitUnknown, err, "failed to init XC")
	}

	kl, err := crypto.ListPrivateKeyIDs(ctx)
	if err != nil {
		return exitError(ctx, ExitUnknown, err, "failed to list private keys")
	}

	out.Print(ctx, "XC Private Keys:")
	for _, key := range kl {
		out.Print(ctx, "%s - %s", key, crypto.FormatKey(ctx, key))
	}

	return nil
}

// XCListPublicKeys lists the XC public keys
func (s *Action) XCListPublicKeys(ctx context.Context, c *cli.Context) error {
	cfgdir := config.Directory()
	crypto, err := xc.New(cfgdir, client.New(cfgdir))
	if err != nil {
		return exitError(ctx, ExitUnknown, err, "failed to init XC")
	}

	kl, err := crypto.ListPublicKeyIDs(ctx)
	if err != nil {
		return exitError(ctx, ExitUnknown, err, "failed to list public keys")
	}

	out.Print(ctx, "XC Public Keys:")
	for _, key := range kl {
		out.Print(ctx, "%s - %s", key, crypto.FormatKey(ctx, key))
	}

	return nil
}

// XCGenerateKeypair generates a new XC keypair
func (s *Action) XCGenerateKeypair(ctx context.Context, c *cli.Context) error {
	name := c.String("name")
	email := c.String("email")
	pw := c.String("passphrase")

	if name == "" {
		var err error
		name, err = termio.AskForString(ctx, "What is your full name?", "")
		if err != nil || name == "" {
			return exitError(ctx, ExitNoName, err, "please provide a name")
		}
	}
	if email == "" {
		var err error
		email, err = termio.AskForString(ctx, "What is your email?", "")
		if err != nil || name == "" {
			return exitError(ctx, ExitNoName, err, "please provide a email")
		}
	}
	if pw == "" {
		var err error
		pw, err = termio.AskForPassword(ctx, name)
		if err != nil {
			return exitError(ctx, ExitIO, err, "failed to ask for password: %s", err)
		}
	}

	cfgdir := config.Directory()
	crypto, err := xc.New(cfgdir, client.New(cfgdir))
	if err != nil {
		return exitError(ctx, ExitUnknown, err, "failed to init XC")
	}

	return crypto.CreatePrivateKeyBatch(ctx, name, email, pw)
}

// XCExportPublicKey exports an XC key
func (s *Action) XCExportPublicKey(ctx context.Context, c *cli.Context) error {
	id := c.String("id")
	file := c.String("file")

	if id == "" {
		return exitError(ctx, ExitUsage, nil, "need id")
	}
	if file == "" {
		return exitError(ctx, ExitUsage, nil, "need file")
	}

	cfgdir := config.Directory()
	crypto, err := xc.New(cfgdir, client.New(cfgdir))
	if err != nil {
		return exitError(ctx, ExitUnknown, err, "failed to init XC")
	}

	if fsutil.IsFile(file) {
		return exitError(ctx, ExitUnknown, nil, "output file already exists")
	}

	pk, err := crypto.ExportPublicKey(ctx, id)
	if err != nil {
		return exitError(ctx, ExitUnknown, err, "failed to export key: %s", err)
	}

	if err := ioutil.WriteFile(file, pk, 0600); err != nil {
		return exitError(ctx, ExitIO, err, "failed to write file")
	}
	return nil
}

// XCImportPublicKey imports an XC key
func (s *Action) XCImportPublicKey(ctx context.Context, c *cli.Context) error {
	file := c.String("file")

	if file == "" {
		return exitError(ctx, ExitUsage, nil, "need file")
	}

	cfgdir := config.Directory()
	crypto, err := xc.New(cfgdir, client.New(cfgdir))
	if err != nil {
		return exitError(ctx, ExitUnknown, err, "failed to init XC")
	}

	if !fsutil.IsFile(file) {
		return exitError(ctx, ExitNotFound, nil, "input file not found")
	}

	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return exitError(ctx, ExitIO, err, "failed to read file")
	}
	return crypto.ImportPublicKey(ctx, buf)
}

// XCRemoveKey removes a key from the keyring
func (s *Action) XCRemoveKey(ctx context.Context, c *cli.Context) error {
	id := c.String("id")

	if id == "" {
		return exitError(ctx, ExitUsage, nil, "need id")
	}

	cfgdir := config.Directory()
	crypto, err := xc.New(cfgdir, client.New(cfgdir))
	if err != nil {
		return exitError(ctx, ExitUnknown, err, "failed to init XC")
	}

	return crypto.RemoveKey(id)
}

// XCExportPrivateKey exports an XC key
func (s *Action) XCExportPrivateKey(ctx context.Context, c *cli.Context) error {
	id := c.String("id")
	file := c.String("file")

	if id == "" {
		return exitError(ctx, ExitUsage, nil, "need id")
	}
	if file == "" {
		return exitError(ctx, ExitUsage, nil, "need file")
	}

	cfgdir := config.Directory()
	crypto, err := xc.New(cfgdir, client.New(cfgdir))
	if err != nil {
		return exitError(ctx, ExitUnknown, err, "failed to init XC")
	}

	if fsutil.IsFile(file) {
		return exitError(ctx, ExitUnknown, nil, "output file already exists")
	}

	pk, err := crypto.ExportPrivateKey(ctx, id)
	if err != nil {
		return exitError(ctx, ExitUnknown, err, "failed to export key: %s", err)
	}

	if err := ioutil.WriteFile(file, pk, 0600); err != nil {
		return exitError(ctx, ExitIO, err, "failed to write file")
	}
	return nil
}

// XCImportPrivateKey imports an XC key
func (s *Action) XCImportPrivateKey(ctx context.Context, c *cli.Context) error {
	file := c.String("file")

	if file == "" {
		return exitError(ctx, ExitUsage, nil, "need file")
	}

	cfgdir := config.Directory()
	crypto, err := xc.New(cfgdir, client.New(cfgdir))
	if err != nil {
		return exitError(ctx, ExitUnknown, err, "failed to init XC")
	}

	if !fsutil.IsFile(file) {
		return exitError(ctx, ExitNotFound, nil, "input file not found")
	}

	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return exitError(ctx, ExitIO, err, "failed to read file")
	}
	return crypto.ImportPrivateKey(ctx, buf)
}
