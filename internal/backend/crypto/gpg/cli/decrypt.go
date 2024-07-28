package cli

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"

	"github.com/gopasspw/gopass/pkg/debug"
)

// Decrypt will try to decrypt the given file.
func (g *GPG) Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, Timeout)
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

	debug.Log("Running %s %+v", cmd.Path, cmd.Args)
	stdout, err := cmd.Output()
	if err != nil {
		debug.Log("Got %+v when running gpg command!", err)
	}

	return stdout, err
}
