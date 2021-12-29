package clipboard

import (
	"context"
	"fmt"
	"os"

	"github.com/atotto/clipboard"
	"github.com/fatih/color"
	"github.com/gopasspw/gopass/internal/notify"
	"github.com/gopasspw/gopass/internal/out"
)

var (
	// Helpers can be overridden at compile time, e.g. go build \
	// -ldflags=='-X github.com/gopasspw/gopass/pkg/clipboard.Helpers=termux-api'.
	Helpers = "xsel or xclip"
	// ErrNotSupported is returned when the clipboard is not accessible.
	ErrNotSupported = fmt.Errorf("WARNING: No clipboard available. Install " + Helpers + " or use -f to print to console")
)

// CopyTo copies the given data to the clipboard and enqueues automatic
// clearing of the clipboard.
func CopyTo(ctx context.Context, name string, content []byte, timeout int) error {
	if clipboard.Unsupported {
		out.Printf(ctx, "%s", ErrNotSupported)
		_ = notify.Notify(ctx, "gopass - clipboard", fmt.Sprintf("%s", ErrNotSupported))
		return nil
	}

	if err := copyToClipboard(ctx, content); err != nil {
		_ = notify.Notify(ctx, "gopass - clipboard", "failed to write to clipboard")
		return fmt.Errorf("failed to write to clipboard: %w", err)
	}

	if timeout < 1 {
		timeout = 45
	}
	if err := clear(ctx, content, timeout); err != nil {
		_ = notify.Notify(ctx, "gopass - clipboard", "failed to clear clipboard")
		return fmt.Errorf("failed to clear clipboard: %w", err)
	}

	out.Printf(ctx, "✔ Copied %s to clipboard. Will clear in %d seconds.", color.YellowString(name), timeout)
	_ = notify.Notify(ctx, "gopass - clipboard", fmt.Sprintf("✔ Copied %s to clipboard. Will clear in %d seconds.", name, timeout))
	return nil
}

func killProc(pid int) {
	// err should be always nil, but just to be sure
	proc, err := os.FindProcess(pid)
	if err != nil {
		return
	}
	// we ignore this error as we're going to return nil anyway
	_ = proc.Kill()
}
