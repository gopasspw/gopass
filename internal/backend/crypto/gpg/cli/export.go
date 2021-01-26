package cli

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/gopasspw/gopass/pkg/debug"
)

// ExportPublicKey will export the named public key to the location given
func (g *GPG) ExportPublicKey(ctx context.Context, id string) ([]byte, error) {
	if id == "" {
		return nil, fmt.Errorf("id is empty")
	}

	args := append(g.args, "--armor", "--export", id)
	cmd := exec.CommandContext(ctx, g.binary, args...)

	debug.Log("%s %+v", cmd.Path, cmd.Args)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run command '%s %+v': %w", cmd.Path, cmd.Args, err)
	}

	if len(out) < 1 {
		return nil, fmt.Errorf("key not found")
	}

	return out, nil
}
