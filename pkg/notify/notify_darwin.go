// +build darwin

package notify

import (
	"context"
	"os"
	"os/exec"

	"github.com/justwatchcom/gopass/pkg/ctxutil"
)

// Notify displays a desktop notification using osascript
func Notify(ctx context.Context, subj, msg string) error {
	if os.Getenv("GOPASS_NO_NOTIFY") != "" || !ctxutil.IsNotifications(ctx) {
		return nil
	}
	osas, err := exec.LookPath("osascript")
	if err != nil {
		return err
	}

	return exec.Command(
		osas,
		"-e",
		`display notification "`+msg+`" with title "`+subj+`"`,
	).Start()
}
