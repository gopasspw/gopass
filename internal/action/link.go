package action

import (
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/urfave/cli/v2"
)

// Link creates a symlink.
func (s *Action) Link(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)

	from := c.Args().Get(0)
	to := c.Args().Get(1)

	if from == "" || to == "" {
		return ExitError(ExitUsage, nil, "Usage: link <from> <to>")
	}

	return s.Store.Link(ctx, from, to)
}
