//go:build darwin
// +build darwin

package notify

import (
	"context"
	"os"
	"os/exec"

	"github.com/gopasspw/gopass/internal/config"
)

const (
	terminalNotifier string = "terminal-notifier"
	osascript        string = "osascript"
)

var (
	execCommand  = exec.Command
	execLookPath = exec.LookPath
)

// Notify displays a desktop notification using osascript.
func Notify(ctx context.Context, subj, msg string) error {
	if os.Getenv("GOPASS_NO_NOTIFY") != "" || !config.Bool(ctx, "core.notifications") {
		return nil
	}

	// check if terminal-notifier was installed else use the applescript fallback
	tn, _ := executableExists(terminalNotifier)
	if tn {
		return tnNotification(ctx, msg, subj)
	}

	return osaNotification(msg, subj)
}

// display notification with osascript.
func osaNotification(msg string, subj string) error {
	_, err := executableExists(osascript)
	if err != nil {
		return err
	}
	args := []string{"-e", `display notification "` + msg + `" with title "` + subj + `"`}

	return execNotification(osascript, args)
}

// exec notification program with passed arguments.
func execNotification(executable string, args []string) error {
	return execCommand(executable, args...).Start()
}

// display notification with terminal-notifier.
func tnNotification(ctx context.Context, msg string, subj string) error {
	arguments := []string{"-title", "Gopass", "-message", msg, "-subtitle", subj, "-appIcon", iconURI(ctx)}

	return execNotification(terminalNotifier, arguments)
}

// check if executable exists.
func executableExists(executable string) (bool, error) {
	_, err := execLookPath(executable)
	if err != nil {
		return false, err
	}

	return true, nil
}
