package cli

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/gopasspw/gopass/pkg/debug"
)

// ImportPublicKey will import a key from the given location
func (g *GPG) ImportPublicKey(ctx context.Context, buf []byte) error {
	if len(buf) < 1 {
		return fmt.Errorf("empty input")
	}

	args := append(g.args, "--import")
	cmd := exec.CommandContext(ctx, g.binary, args...)
	cmd.Stdin = bytes.NewReader(buf)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	debug.Log("gpg.ImportPublicKey: %s %+v", cmd.Path, cmd.Args)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run command: '%s %+v': %w", cmd.Path, cmd.Args, err)
	}

	// clear key cache
	g.privKeys = nil
	g.pubKeys = nil
	return nil
}
