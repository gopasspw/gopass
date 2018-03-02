// +build windows

package notify

import (
	"context"
	"os"
	"os/exec"

	"github.com/justwatchcom/gopass/utils/ctxutil"
)

// Notify displays a desktop notification through msg
func Notify(ctx context.Context, subj, msg string) error {
	if os.Getenv("GOPASS_NO_NOTIFY") != "" || ctxutil.IsNotifications(ctx) {
		return nil
	}
	winmsg, err := exec.LookPath("msg")
	if err != nil {
		return err
	}

	return exec.Command(winmsg,
		"*",
		"/TIME:3",
		subj+"\n\n"+msg,
	).Start()
}
