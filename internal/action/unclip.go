package action

import (
	"os"
	"time"

	"github.com/gopasspw/gopass/internal/action/exit"
	"github.com/gopasspw/gopass/pkg/clipboard"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/urfave/cli/v2"
)

// Unclip tries to erase the content of the clipboard.
func (s *Action) Unclip(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	force := c.Bool("force")
	timeout := c.Int("timeout")
	name := os.Getenv("GOPASS_UNCLIP_NAME")
	checksum := os.Getenv("GOPASS_UNCLIP_CHECKSUM")

	time.Sleep(time.Second * time.Duration(timeout))
	if err := clipboard.Clear(ctx, name, checksum, force); err != nil {
		return exit.Error(exit.IO, err, "Failed to clear clipboard: %s", err)
	}

	return nil
}
