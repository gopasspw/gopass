// +build darwin

package notify

import (
	"os"
	"os/exec"
)

// Notify displays a desktop notification using osascript
func Notify(subj, msg string) error {
	if nn := os.Getenv("GOPASS_NO_NOTIFY"); nn != "" {
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
