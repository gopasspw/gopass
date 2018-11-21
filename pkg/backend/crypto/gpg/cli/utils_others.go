// +build !linux,!windows

package cli

import (
	"os"
	"os/exec"
)

func tty() string {
	cmd := exec.Command("/usr/bin/tty")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(out)
}
