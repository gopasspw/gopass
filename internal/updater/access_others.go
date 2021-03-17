// +build !windows

package updater

import "golang.org/x/sys/unix"

func canWrite(path string) error {
	return unix.Access(path, unix.W_OK)
}
