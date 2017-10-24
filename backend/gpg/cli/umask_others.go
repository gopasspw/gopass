// +build !windows

package cli

import "syscall"

func umask(mask int) int {
	return syscall.Umask(mask)
}
