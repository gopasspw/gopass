package action

import (
	"context"
	"os"
	"time"

	"github.com/gopasspw/gopass/pkg/clipboard"

	"gopkg.in/urfave/cli.v1"
)

// Unclip tries to erase the content of the clipboard
func (s *Action) Unclip(ctx context.Context, c *cli.Context) error {
	force := c.Bool("force")
	timeout := c.Int("timeout")
	checksum := os.Getenv("GOPASS_UNCLIP_CHECKSUM")

	time.Sleep(time.Second * time.Duration(timeout))
	if err := clipboard.Clear(ctx, checksum, force); err != nil {
		return ExitError(ctx, ExitIO, err, "Failed to clear clipboard: %s", err)
	}
	return nil
}
