package action

import (
	"context"
	"fmt"

	"github.com/justwatchcom/gopass/utils/termio"
	"github.com/urfave/cli"
)

// Move the content from one secret to another
func (s *Action) Move(ctx context.Context, c *cli.Context) error {
	force := c.Bool("force")

	if len(c.Args()) != 2 {
		return exitError(ctx, ExitUsage, nil, "Usage: %s mv old-path new-path", s.Name)
	}

	from := c.Args()[0]
	to := c.Args()[1]

	if !force {
		if s.Store.Exists(ctx, to) && !termio.AskForConfirmation(ctx, fmt.Sprintf("%s already exists. Overwrite it?", to)) {
			return exitError(ctx, ExitAborted, nil, "not overwriting your current secret")
		}
	}

	if err := s.Store.Move(ctx, from, to); err != nil {
		return exitError(ctx, ExitUnknown, err, "%s", err)
	}

	return nil
}
