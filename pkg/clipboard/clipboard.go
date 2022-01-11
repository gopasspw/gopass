package clipboard

import (
	"context"
	"fmt"
	"os"
	"os/exec"

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
	ErrNotSupported = fmt.Errorf("WARNING: No clipboard available. Install " + Helpers + ", provide $GOPASS_CLIPBOARD_COPY_CMD and $GOPASS_CLIPBOARD_CLEAR_CMD or use -f to print to console")
)

// CopyTo copies the given data to the clipboard and enqueues automatic
// clearing of the clipboard.
func CopyTo(ctx context.Context, name string, content []byte, timeout int) error {
	clipboardCopyCMD := os.Getenv("GOPASS_CLIPBOARD_COPY_CMD")
	if clipboardCopyCMD != "" {
		if err := callCommand(ctx, clipboardCopyCMD, name, content); err != nil {
			_ = notify.Notify(ctx, "gopass - clipboard", "failed to call clipboard copy command")
			return fmt.Errorf("failed to call clipboard copy command: %w", err)
		}
	} else if clipboard.Unsupported {
		out.Errorf(ctx, "%s", ErrNotSupported)
		_ = notify.Notify(ctx, "gopass - clipboard", fmt.Sprintf("%s", ErrNotSupported))
		return nil
	} else if err := copyToClipboard(ctx, content); err != nil {
		_ = notify.Notify(ctx, "gopass - clipboard", "failed to write to clipboard")
		return fmt.Errorf("failed to write to clipboard: %w", err)
	}

	if timeout < 1 {
		timeout = 45
	}
	if err := clear(ctx, name, content, timeout); err != nil {
		_ = notify.Notify(ctx, "gopass - clipboard", "failed to clear clipboard")
		return fmt.Errorf("failed to clear clipboard: %w", err)
	}

	out.Printf(ctx, "✔ Copied %s to clipboard. Will clear in %d seconds.", color.YellowString(name), timeout)
	_ = notify.Notify(ctx, "gopass - clipboard", fmt.Sprintf("✔ Copied %s to clipboard. Will clear in %d seconds.", name, timeout))
	return nil
}

func callCommand(ctx context.Context, cmd string, parameter string, stdinValue []byte) error {
	clipboardProcess := exec.Command(cmd, parameter)
	stdin, err := clipboardProcess.StdinPipe()
	defer stdin.Close()

	if err != nil {
		return err
	}
	if err = clipboardProcess.Start(); err != nil {
		return err
	}
	if _, err = stdin.Write(stdinValue); err != nil {
		return err
	}

	// Force STDIN close before we wait for the process to finish, so we avoid deadlocks
	if err = stdin.Close(); err != nil {
		return err
	}

	if err := clipboardProcess.Wait(); err != nil {
		return err
	}
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
