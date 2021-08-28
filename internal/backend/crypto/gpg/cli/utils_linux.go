//go:build linux
// +build linux

package cli

import (
	"os"
	"syscall"
)

var (
	fd0 = "/proc/self/fd/0"
)

// see https://www.gnupg.org/documentation/manuals/gnupg/Invoking-GPG_002dAGENT.html
func tty() string {
	dest, err := os.Readlink(fd0)
	if err != nil {
		return ""
	}
	return dest
}

func umask(mask int) int {
	return syscall.Umask(mask)
}
