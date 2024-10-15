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
	"github.com/gopasspw/gopass/pkg/debug"
)

var (
	// Helpers can be overridden at compile time, e.g. go build \
	// -ldflags=='-X github.com/gopasspw/gopass/pkg/clipboard.Helpers=termux-api'.
	Helpers = "xsel or xclip"
	// ErrNotSupported is returned when the clipboard is not accessible.
	ErrNotSupported = fmt.Errorf("WARNING: No clipboard available. "+
		"Install %s, provide $GOPASS_CLIPBOARD_COPY_CMD and $GOPASS_CLIPBOARD_CLEAR_CMD or use -f to print to console", Helpers)
)

// CopyTo copies the given data to the clipboard and enqueues automatic
// clearing of the clipboard.
func CopyTo(ctx context.Context, name string, content []byte, timeout int) error {
	debug.Log("Copying to clipboard: %s for %ds", name, timeout)

	clipboardCopyCMD := os.Getenv("GOPASS_CLIPBOARD_COPY_CMD")
	if clipboardCopyCMD != "" {
		if err := callCommand(ctx, clipboardCopyCMD, name, content); err != nil {
			_ = notify.Notify(ctx, "gopass - clipboard", "failed to call clipboard copy command")

			return fmt.Errorf("failed to call clipboard copy command: %w", err)
		}
	} else if clipboard.Unsupported {
		out.Errorf(ctx, "%s", ErrNotSupported)
		_ = notify.Notify(ctx, "gopass - clipboard", ErrNotSupported.Error())

		return nil
	} else if err := copyToClipboard(ctx, content); err != nil {
		_ = notify.Notify(ctx, "gopass - clipboard", "failed to write to clipboard")

		return fmt.Errorf("failed to write to clipboard: %w", err)
	}

	if timeout < 1 {
		debug.Log("Auto-clear of clipboard disabled.")

		out.Printf(ctx, "✔ Copied %s to clipboard.", color.YellowString(name))
		_ = notify.Notify(ctx, "gopass - clipboard", fmt.Sprintf("✔ Copied %s to clipboard.", name))

		return nil
	}

	if err := clearClip(ctx, name, content, timeout); err != nil {
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

	defer func() {
		_ = stdin.Close()
	}()

	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	if err = clipboardProcess.Start(); err != nil {
		return fmt.Errorf("failed to start clipboard process: %w", err)
	}

	if _, err = stdin.Write(stdinValue); err != nil {
		return fmt.Errorf("failed to write to STDIN: %w", err)
	}

	// Force STDIN close before we wait for the process to finish, so we avoid deadlocks
	if err = stdin.Close(); err != nil {
		return fmt.Errorf("failed to close STDIN: %w", err)
	}

	if err := clipboardProcess.Wait(); err != nil {
		return fmt.Errorf("failed to call clipboard command: %w", err)
	}

	return nil
}

func killProc(pid int) {
	// err should be always nil, but just to be sure
	proc, err := os.FindProcess(pid)
	if err != nil {
		return
	}

	if err := proc.Kill(); err != nil {
		debug.Log("failed to kill %d: %s", pid, err)

		return
	}

	// wait for the process to actually exit to avoid zombie processes
	ps, err := proc.Wait()
	if err != nil {
		debug.Log("failed to wait for %d: %s", pid, err)

		return
	}

	debug.Log("killed process exited with %d", ps.ExitCode())
}
