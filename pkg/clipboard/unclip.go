package clipboard

import (
	"context"
	"fmt"
	"os"

	"github.com/atotto/clipboard"
	"github.com/gopasspw/gopass/internal/notify"
	"github.com/gopasspw/gopass/internal/pwschemes/argon2id"
	"github.com/gopasspw/gopass/pkg/debug"
)

// Clear will attempt to erase the clipboard.
func Clear(ctx context.Context, name string, checksum string, force bool) error {
	clipboardClearCMD := os.Getenv("GOPASS_CLIPBOARD_CLEAR_CMD")
	if clipboardClearCMD != "" {
		if err := callCommand(ctx, clipboardClearCMD, name, []byte(checksum)); err != nil {
			_ = notify.Notify(ctx, "gopass - clipboard", "failed to call clipboard clear command")
			return fmt.Errorf("failed to call clipboard clear command: %w", err)
		}
		debug.Log("clipboard cleared (%s)", checksum)
		return nil
	}

	if clipboard.Unsupported {
		return ErrNotSupported
	}

	cur, err := clipboard.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read clipboard: %w", err)
	}

	match, err := argon2id.Validate(cur, checksum)
	if err != nil {
		debug.Log("failed to validate checksum %s: %s", checksum, err)
		return nil
	}
	if !match && !force {
		return nil
	}

	if err := clipboard.WriteAll(""); err != nil {
		_ = notify.Notify(ctx, "gopass - clipboard", "Failed to clear clipboard")
		return fmt.Errorf("failed to write clipboard: %w", err)
	}

	if err := clearClipboardHistory(ctx); err != nil {
		_ = notify.Notify(ctx, "gopass - clipboard", "Failed to clear clipboard history")
		return fmt.Errorf("failed to clear clipboard history: %w", err)
	}

	if err := notify.Notify(ctx, "gopass - clipboard", "Clipboard has been cleared"); err != nil {
		return fmt.Errorf("failed to send unclip notification: %w", err)
	}

	debug.Log("clipboard cleared (%s)", checksum)
	return nil
}
