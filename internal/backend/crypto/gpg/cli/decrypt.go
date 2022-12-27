package cli

import (
	"bytes"
	"context"
	"os"
	"os/exec"

	"github.com/gopasspw/gopass/pkg/debug"
)

// Decrypt will try to decrypt the given file.
func (g *GPG) Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()

	args := append(g.args, "--decrypt")
	cmd := exec.CommandContext(ctx, g.binary, args...)
	cmd.Stdin = bytes.NewReader(ciphertext)
	cmd.Stderr = os.Stderr

	debug.Log("%s %+v", cmd.Path, cmd.Args)

	return cmd.Output()
}
