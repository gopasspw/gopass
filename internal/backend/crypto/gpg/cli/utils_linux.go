// +build linux

package cli

import "os"

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
