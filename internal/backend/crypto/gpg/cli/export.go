package cli

import (
	"context"
	"os/exec"

	"github.com/gopasspw/gopass/internal/debug"
	"github.com/pkg/errors"
)

// ExportPublicKey will export the named public key to the location given
func (g *GPG) ExportPublicKey(ctx context.Context, id string) ([]byte, error) {
	if id == "" {
		return nil, errors.Errorf("id is empty")
	}

	args := append(g.args, "--armor", "--export", id)
	cmd := exec.CommandContext(ctx, g.binary, args...)

	debug.Log("%s %+v", cmd.Path, cmd.Args)
	out, err := cmd.Output()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to run command '%s %+v'", cmd.Path, cmd.Args)
	}

	if len(out) < 1 {
		return nil, errors.Errorf("Key not found")
	}

	return out, nil
}
