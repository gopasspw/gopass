package cli

import (
	"bytes"
	"context"
	"os"
	"os/exec"

	"github.com/gopasspw/gopass/internal/out"

	"github.com/pkg/errors"
)

// ImportPublicKey will import a key from the given location
func (g *GPG) ImportPublicKey(ctx context.Context, buf []byte) error {
	if len(buf) < 1 {
		return errors.Errorf("empty input")
	}

	args := append(g.args, "--import")
	cmd := exec.CommandContext(ctx, g.binary, args...)
	cmd.Stdin = bytes.NewReader(buf)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	out.Debug(ctx, "gpg.ImportPublicKey: %s %+v", cmd.Path, cmd.Args)
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "failed to run command: '%s %+v'", cmd.Path, cmd.Args)
	}

	// clear key cache
	g.privKeys = nil
	g.pubKeys = nil
	return nil
}
