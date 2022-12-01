//go:build !windows
// +build !windows

package updater

import "golang.org/x/sys/unix"

func canWrite(path string) error {
	return unix.Access(path, unix.W_OK) //nolint:wrapcheck
}

func removeOldBinary(dir, dest string) error {
	// no need, os.Rename will replace the destination
	return nil
}
