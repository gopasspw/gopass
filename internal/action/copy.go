package action

import (
	"context"
	"fmt"

	"github.com/gopasspw/gopass/internal/action/exit"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/termio"
	"github.com/urfave/cli/v2"
)

// Copy the contents of a file to another one.
func (s *Action) Copy(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	force := c.Bool("force")

	if c.Args().Len() != 2 {
		return exit.Error(exit.Usage, nil, "Usage: %s cp <FROM> <TO>", s.Name)
	}

	from := c.Args().Get(0)
	to := c.Args().Get(1)

	return s.copy(ctx, from, to, force)
}

func (s *Action) copy(ctx context.Context, from, to string, force bool) error {
	if !s.Store.Exists(ctx, from) && !s.Store.IsDir(ctx, from) {
		return exit.Error(exit.NotFound, nil, "%s does not exist", from)
	}

	if !force {
		if s.Store.Exists(ctx, to) && !termio.AskForConfirmation(ctx, fmt.Sprintf("%s already exists. Overwrite it?", to)) {
			return exit.Error(exit.Aborted, nil, "not overwriting your current secret")
		}
	}

	if err := s.Store.Copy(ctx, from, to); err != nil {
		return exit.Error(exit.IO, err, "failed to copy from %q to %q", from, to)
	}

	return nil
}
