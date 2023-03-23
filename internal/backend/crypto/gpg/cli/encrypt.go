package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/gopasspw/gopass/internal/backend/crypto/gpg"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/debug"
)

// Encrypt will encrypt the given content for the recipients. If alwaysTrust is true
// the trust-model will be set to always as to avoid (annoying) "unusable public key"
// errors when encrypting.
func (g *GPG) Encrypt(ctx context.Context, plaintext []byte, recipients []string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()

	args := append(g.args, "--encrypt")
	if gpg.IsAlwaysTrust(ctx) {
		// changing the trustmodel is possibly dangerous. A user should always
		// explicitly opt-in to do this
		args = append(args, "--trust-model=always")
	}

	buf := &bytes.Buffer{}
	if len(recipients) == 0 {
		return buf.Bytes(), errors.New("recipients list is empty!")
	}
	var badRecipients []string
	for _, r := range recipients {
		kl, err := g.listKeys(ctx, "public", r)
		if err != nil {
			debug.Log("Failed to check key %s. Adding anyway. %s", err)
		} else if len(kl.UseableKeys(gpg.IsAlwaysTrust(ctx))) < 1 {
			badRecipients = append(badRecipients, r)
			errmsg := fmt.Sprintf("Not using invalid key %q for encryption. Check its expiration date, its encryption capabilities and trust.", r)
			debug.Log(errmsg)
			out.Printf(ctx, errmsg)

			continue
		}
		debug.Log("adding recipient %s", r)
		args = append(args, "--recipient", r)
	}
	if len(badRecipients) == len(recipients) {
		return buf.Bytes(), errors.New("no valid and trusted recipients were found!")
	}

	cmd := exec.CommandContext(ctx, g.binary, args...)
	cmd.Stdin = bytes.NewReader(plaintext)
	// the encrypted blob and errors are printed to the log file, and to stdout
	cmd.Stdout = io.MultiWriter(buf, debug.LogWriter)
	cmd.Stderr = io.MultiWriter(os.Stderr, debug.LogWriter)

	debug.Log("%s %+v", cmd.Path, cmd.Args)
	err := cmd.Run()

	return buf.Bytes(), err
}
