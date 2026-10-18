package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/gopasspw/gopass/pkg/debug"
)

// Decrypt will try to decrypt the given file.
func (g *GPG) Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error) {
	timeout := timeoutFromContext(ctx)
	ctx, cancel := withTimeout(ctx, timeout)
	defer cancel()

	args := append(g.args, "--decrypt")
	// Useful information may appear there
	if debug.IsEnabled() {
		args = append(args, "--verbose", "--verbose")
	}
	cmd := exec.CommandContext(ctx, g.binary, args...)
	cmd.Stdin = bytes.NewReader(ciphertext)
	// If gopass-jsonapi is used, there is no way to reach this os.Stderr, so
	// we write this stderr to the log file as well.
	cmd.Stderr = io.MultiWriter(os.Stderr, debug.LogWriter)

	debug.V(1).Log("Running %s %+v", cmd.Path, cmd.Args)
	stdout, err := cmd.Output()
	if err != nil {
		debug.Log("GPG decrypt failed: %s %+v: %+v", cmd.Path, cmd.Args, err)
		if timeout > 0 && errors.Is(ctx.Err(), context.DeadlineExceeded) {
			// GPG was killed because it did not finish within the deadline.
			// Most often that is a passphrase prompt nobody answered in time,
			// so point at the knob that governs the wait.
			err = fmt.Errorf("gpg decrypt did not finish within %s. If you need more time to enter your GPG passphrase, raise the timeout with `gopass config gpg.timeout <seconds>`, e.g. `gopass config gpg.timeout 600` for ten minutes, or set it to 0 to disable the deadline: %w", timeout, err)
		}
	}

	return stdout, err
}
