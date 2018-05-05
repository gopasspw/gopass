package clipboard

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/justwatchcom/gopass/pkg/notify"

	"github.com/atotto/clipboard"
	"github.com/pkg/errors"
)

// Clear will attempt to erase the clipboard
func Clear(ctx context.Context, checksum string, force bool) error {
	if clipboard.Unsupported {
		return ErrNotSupported
	}

	cur, err := clipboard.ReadAll()
	if err != nil {
		return errors.Wrapf(err, "failed to read clipboard: %s", err)
	}

	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(cur)))
	if hash != checksum && !force {
		return nil
	}

	if err := clipboard.WriteAll(""); err != nil {
		_ = notify.Notify(ctx, "gopass - clipboard", "Failed to clear clipboard")
		return errors.Wrapf(err, "failed to write clipboard: %s", err)
	}

	if err := clearClipboardHistory(ctx); err != nil {
		_ = notify.Notify(ctx, "gopass - clipboard", "Failed to clear clipboard history")
		return errors.Wrapf(err, "failed to clear clipboard history: %s", err)
	}

	if err := notify.Notify(ctx, "gopass -clipboard", "Clipboard has been cleared"); err != nil {
		return errors.Wrapf(err, "failed to send unclip notification: %s", err)
	}

	return nil
}
