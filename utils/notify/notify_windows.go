// +build windows

package notify

import (
	"os"
	"os/exec"
)

// Notify displays a desktop notification through msg
func Notify(subj, msg string) error {
	if nn := os.Getenv("GOPASS_NO_NOTIFY"); nn != "" {
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
