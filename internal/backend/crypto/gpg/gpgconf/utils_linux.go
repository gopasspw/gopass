//go:build linux

package gpgconf

import (
	"os"
	"syscall"
)

var fd0 = "/proc/self/fd/0"

// TTY returns the tty of the current process.
// see https://www.gnupg.org/documentation/manuals/gnupg/Invoking-GPG_002dAGENT.html
func TTY() string {
	dest, err := os.Readlink(fd0)
	if err != nil {
		return ""
	}

	return dest
}

// Umask sets the desired umask.
func Umask(mask int) int {
	return syscall.Umask(mask)
}
