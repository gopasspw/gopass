package action

import (
	"context"

	"github.com/justwatchcom/gopass/pkg/out"
	"github.com/justwatchcom/gopass/pkg/updater"
	"github.com/urfave/cli"
)

// Update will start the interactive update assistant
func (s *Action) Update(ctx context.Context, c *cli.Context) error {
	pre := c.Bool("pre")

	if s.version.String() == "0.0.0+HEAD" {
		out.Red(ctx, "Can not check version against HEAD")
		return nil
	}

	if err := updater.Update(ctx, pre, s.version); err != nil {
		return ExitError(ctx, ExitUnknown, err, "Failed to update gopass: %s", err)
	}
	return nil
}
