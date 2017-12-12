// +build darwin

package notify

import "os/exec"

// Notify displays a desktop notification using osascript
func Notify(subj, msg string) error {
	osas, err := exec.LookPath("osascript")
	if err != nil {
		return err
	}

	return exec.Command(
		osa,
		"-e",
		`display notification "`+msg+`" with title "`+subj+`"`,
	).Start()
}
