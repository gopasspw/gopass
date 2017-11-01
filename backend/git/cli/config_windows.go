// +build windows

package cli

import (
	"context"

	"github.com/pkg/errors"
)

func (g *Git) fixConfigOSDep(ctx context.Context) error {
	if g.gpg == "" {
		return nil
	}

	if err := g.Cmd(ctx, "gitFixConfigOSDep", "config", "--local", "gpg.program", g.gpg); err != nil {
		return errors.Wrapf(err, "failed to set git config gpg.program")
	}
	return nil
}
