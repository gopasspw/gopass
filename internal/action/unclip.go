package action

import (
	"os"
	"time"

	"github.com/gopasspw/gopass/pkg/clipboard"
	"github.com/gopasspw/gopass/pkg/ctxutil"

	"github.com/urfave/cli/v2"
)

// Unclip tries to erase the content of the clipboard
func (s *Action) Unclip(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	force := c.Bool("force")
	timeout := c.Int("timeout")
	checksum := os.Getenv("GOPASS_UNCLIP_CHECKSUM")

	time.Sleep(time.Second * time.Duration(timeout))
	if err := clipboard.Clear(ctx, checksum, force); err != nil {
		return ExitError(ExitIO, err, "Failed to clear clipboard: %s", err)
	}
	return nil
}
