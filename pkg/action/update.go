package action

import (
	"context"

	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/pkg/updater"

	"gopkg.in/urfave/cli.v1"
)

// Update will start the interactive update assistant
func (s *Action) Update(ctx context.Context, c *cli.Context) error {
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
