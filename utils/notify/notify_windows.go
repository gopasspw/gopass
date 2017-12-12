// +build windows

package notify

import "os/exec"

// Notify displays a desktop notification through msg
func Notify(subj, msg string) error {
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
