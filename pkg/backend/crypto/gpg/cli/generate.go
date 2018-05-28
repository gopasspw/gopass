package cli

import (
	"bytes"
	"context"
	"os"
	"os/exec"

	"github.com/gopasspw/gopass/pkg/out"

	"github.com/pkg/errors"
)

// CreatePrivateKeyBatch will create a new GPG keypair in batch mode
func (g *GPG) CreatePrivateKeyBatch(ctx context.Context, name, email, passphrase string) error {
	buf := &bytes.Buffer{}
	// https://git.gnupg.org/cgi-bin/gitweb.cgi?p=gnupg.git;a=blob;f=doc/DETAILS;h=de0f21ccba60c3037c2a155156202df1cd098507;hb=refs/heads/STABLE-BRANCH-1-4#l716
	_, _ = buf.WriteString(`%echo Generating a RSA/RSA key pair
Key-Type: RSA
Key-Length: 2048
Subkey-Type: RSA
Subkey-Length: 2048
Expire-Date: 0
`)
	_, _ = buf.WriteString("Name-Real: " + name + "\n")
	_, _ = buf.WriteString("Name-Email: " + email + "\n")
	_, _ = buf.WriteString("Passphrase: " + passphrase + "\n")

	args := []string{"--batch", "--gen-key"}
	cmd := exec.CommandContext(ctx, g.binary, args...)
	cmd.Stdin = bytes.NewReader(buf.Bytes())
	cmd.Stdout = nil
	cmd.Stderr = nil

	out.Debug(ctx, "gpg.CreatePrivateKeyBatch: %s %+v", cmd.Path, cmd.Args)
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "failed to run command: '%s %+v'", cmd.Path, cmd.Args)
	}
	g.privKeys = nil
	g.pubKeys = nil
	return nil
}

// CreatePrivateKey will create a new GPG key in interactive mode
func (g *GPG) CreatePrivateKey(ctx context.Context) error {
	args := []string{"--gen-key"}
	cmd := exec.CommandContext(ctx, g.binary, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	out.Debug(ctx, "gpg.CreatePrivateKey: %s %+v", cmd.Path, cmd.Args)
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "failed to run command: '%s %+v'", cmd.Path, cmd.Args)
	}

	g.privKeys = nil
	g.pubKeys = nil
	return nil
}
