package action

import (
	"fmt"

	"github.com/gopasspw/gopass/internal/termio"
	"github.com/gopasspw/gopass/pkg/ctxutil"

	"github.com/urfave/cli/v2"
)

// Move the content from one secret to another
func (s *Action) Move(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	force := c.Bool("force")

	if c.Args().Len() != 2 {
		return ExitError(ExitUsage, nil, "Usage: %s mv old-path new-path", s.Name)
	}

	from := c.Args().Get(0)
	to := c.Args().Get(1)

	if !force {
		if s.Store.Exists(ctx, to) && !termio.AskForConfirmation(ctx, fmt.Sprintf("%s already exists. Overwrite it?", to)) {
			return ExitError(ExitAborted, nil, "not overwriting your current secret")
		}
	}

	if err := s.Store.Move(ctx, from, to); err != nil {
		return ExitError(ExitUnknown, err, "%s", err)
	}

	return nil
}
