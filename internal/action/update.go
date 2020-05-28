package action

import (
	"context"

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

	// migration check is not yet implemented. returning false will bock
	// updates to the next major release.
	mc := func(ctx context.Context) bool {
		return false
	}

	if err := updater.Update(ctx, pre, s.version, mc); err != nil {
		return ExitError(ExitUnknown, err, "Failed to update gopass: %s", err)
	}
	return nil
}
