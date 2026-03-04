//go:build !linux && !windows

package gpgconf

import (
	"os"
	"os/exec"
	"syscall"
)

func TTY() string {
	cmd := exec.Command("/usr/bin/tty")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return ""
	}

	return string(out)
}

func Umask(mask int) int {
	return syscall.Umask(mask)
}
