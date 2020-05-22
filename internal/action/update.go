package action

import (
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/updater"
	"github.com/gopasspw/gopass/pkg/ctxutil"

	"github.com/urfave/cli/v2"
)

// Update will start the interactive update assistant
func (s *Action) Update(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	pre := c.Bool("pre")

	if s.version.String() == "0.0.0+HEAD" {
		out.Error(ctx, "Can not check version against HEAD")
		return nil
	}

	if err := updater.Update(ctx, pre, s.version); err != nil {
		return ExitError(ctx, ExitUnknown, err, "Failed to update gopass: %s", err)
	}
	return nil
}
