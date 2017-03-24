// +build !windows

package gpg

import "syscall"

func umask(mask int) int {
	return syscall.Umask(mask)
}
