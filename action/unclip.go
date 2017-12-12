package action

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"time"

	"github.com/atotto/clipboard"
	"github.com/justwatchcom/gopass/utils/notify"
	"github.com/urfave/cli"
)

// Unclip tries to erase the content of the clipboard
func (s *Action) Unclip(ctx context.Context, c *cli.Context) error {
	force := c.Bool("force")
	timeout := c.Int("timeout")
	checksum := os.Getenv("GOPASS_UNCLIP_CHECKSUM")

	time.Sleep(time.Second * time.Duration(timeout))

	cur, err := clipboard.ReadAll()
	if err != nil {
		return s.exitError(ctx, ExitIO, err, "failed to read clipboard: %s", err)
	}

	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(cur)))

	if hash != checksum && !force {
		return nil
	}

	if err := clipboard.WriteAll(""); err != nil {
		_ = notify.Notify("gopass - clipboard", "Failed to clear clipboard")
		return s.exitError(ctx, ExitIO, err, "failed to write clipboard: %s", err)
	}

	if err := s.clearClipboardHistory(ctx); err != nil {
		_ = notify.Notify("gopass - clipboard", "Failed to clear clipboard history")
		return s.exitError(ctx, ExitIO, err, "failed to clear clipboard history: %s", err)
	}

	if err := notify.Notify("gopass -clipboard", "Clipboard has been cleared"); err != nil {
		return s.exitError(ctx, ExitIO, err, "failed to send unclip notification: %s", err)
	}

	return nil
}
