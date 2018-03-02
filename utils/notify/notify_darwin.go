// +build darwin

package notify

import (
	"context"
	"os"
	"os/exec"

	"github.com/justwatchcom/gopass/utils/ctxutil"
)

// Notify displays a desktop notification using osascript
func Notify(ctx context.Context, subj, msg string) error {
	if os.Getenv("GOPASS_NO_NOTIFY") != "" || ctxutil.IsNotify(ctx) != true {
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
