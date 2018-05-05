package action

import (
	"context"
	"fmt"

	"github.com/justwatchcom/gopass/pkg/termio"

	"github.com/urfave/cli"
)

// Copy the contents of a file to another one
func (s *Action) Copy(ctx context.Context, c *cli.Context) error {
	force := c.Bool("force")

	if len(c.Args()) != 2 {
		return ExitError(ctx, ExitUsage, nil, "Usage: %s cp <FROM> <TO>", s.Name)
	}

	from := c.Args()[0]
	to := c.Args()[1]

	return s.copy(ctx, from, to, force)
}

func (s *Action) copy(ctx context.Context, from, to string, force bool) error {
	if !s.Store.Exists(ctx, from) && !s.Store.IsDir(ctx, from) {
		return ExitError(ctx, ExitNotFound, nil, "%s does not exist", from)
	}

	if !force {
		if s.Store.Exists(ctx, to) && !termio.AskForConfirmation(ctx, fmt.Sprintf("%s already exists. Overwrite it?", to)) {
			return ExitError(ctx, ExitAborted, nil, "not overwriting your current secret")
		}
	}

	if err := s.Store.Copy(ctx, from, to); err != nil {
		return ExitError(ctx, ExitIO, err, "failed to copy from '%s' to '%s'", from, to)
	}

	return nil
}
