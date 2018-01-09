package cli

import (
	"context"
	"os/exec"

	"github.com/justwatchcom/gopass/utils/out"
	"github.com/pkg/errors"
)

// ExportPublicKey will export the named public key to the location given
func (g *GPG) ExportPublicKey(ctx context.Context, id string) ([]byte, error) {
	args := append(g.args, "--armor", "--export", id)
	cmd := exec.CommandContext(ctx, g.binary, args...)

	out.Debug(ctx, "gpg.ExportPublicKey: %s %+v", cmd.Path, cmd.Args)
	out, err := cmd.Output()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to run command '%s %+v'", cmd.Path, cmd.Args)
	}

	if len(out) < 1 {
		return nil, errors.Errorf("Key not found")
	}

	return out, nil
}
