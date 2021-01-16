package action

import (
	"fmt"
	"runtime"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/updater"
	"github.com/gopasspw/gopass/pkg/ctxutil"

	"github.com/urfave/cli/v2"
)

// Update will start the interactive update assistant
func (s *Action) Update(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)

	if s.version.String() == "0.0.0+HEAD" {
		out.Error(ctx, "Can not check version against HEAD")
		return nil
	}

	if runtime.GOOS == "windows" {
		return fmt.Errorf("gopass update is not supported on windows (#1722)")
	}

	out.Print(ctx, "⚒ Checking for available updates ...")
	if err := updater.Update(ctx, s.version); err != nil {
		return ExitError(ExitUnknown, err, "Failed to update gopass: %s", err)
	}
	out.Print(ctx, "✅ gopass is up to date")
	return nil
}
