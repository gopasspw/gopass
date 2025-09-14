package cli

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"

	"github.com/gopasspw/gopass/pkg/debug"
)

var (
	gpgRevocsRE = regexp.MustCompile(`.*/openpgp-revocs.d/([0-9A-F]{40})\.rev`)
)

// GenerateIdentity will create a new GPG keypair in batch mode.
func (g *GPG) GenerateIdentity(ctx context.Context, name, email, passphrase string) (string, error) {
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

	out := &bytes.Buffer{}
	cmd.Stdout = out
	cmd.Stderr = out

	debug.Log("%s %+v", cmd.Path, cmd.Args)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to run command: '%s %+v': %q - %w", cmd.Path, cmd.Args, out.String(), err)
	}

	g.privKeys = nil
	g.pubKeys = nil

	// try to parse key id from the output
	for line := range bytes.SplitSeq(out.Bytes(), []byte{'\n'}) {
		if !gpgRevocsRE.Match(line) {

			continue
		}

		matches := gpgRevocsRE.FindSubmatch(line)
		if len(matches) == 2 {
			keyID := string(matches[1])
			debug.Log("Generated new GPG key: %s", keyID)
			return keyID, nil
		}
	}

	// Ignoring this failure since we're usually not using the Key ID directly.
	debug.Log("Failed to find key ID in output: %q", out.String())

	return "", nil
}
