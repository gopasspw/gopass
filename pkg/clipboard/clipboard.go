package clipboard

import (
	"context"
	"fmt"
	"os"

	"github.com/gopasspw/gopass/internal/notify"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"

	"github.com/atotto/clipboard"
	"github.com/fatih/color"
)

var (
	// ErrNotSupported is returned when the clipboard is not accessible
	ErrNotSupported = fmt.Errorf("WARNING: No clipboard available. Install xsel or xclip or use -f to print to console")
)

// CopyTo copies the given data to the clipboard and enqueues automatic
// clearing of the clipboard
func CopyTo(ctx context.Context, name string, content []byte) error {
	if clipboard.Unsupported {
		out.Printf(ctx, "%s", ErrNotSupported)
		_ = notify.Notify(ctx, "gopass - clipboard", fmt.Sprintf("%s", ErrNotSupported))
		return nil
	}

	if err := clipboard.WriteAll(string(content)); err != nil {
		_ = notify.Notify(ctx, "gopass - clipboard", "failed to write to clipboard")
		return fmt.Errorf("failed to write to clipboard: %w", err)
	}

	if err := clear(ctx, content, ctxutil.GetClipTimeout(ctx)); err != nil {
		_ = notify.Notify(ctx, "gopass - clipboard", "failed to clear clipboard")
		return fmt.Errorf("failed to clear clipboard: %w", err)
	}

	out.Printf(ctx, "✔ Copied %s to clipboard. Will clear in %d seconds.", color.YellowString(name), ctxutil.GetClipTimeout(ctx))
	_ = notify.Notify(ctx, "gopass - clipboard", fmt.Sprintf("✔ Copied %s to clipboard. Will clear in %d seconds.", name, ctxutil.GetClipTimeout(ctx)))
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
