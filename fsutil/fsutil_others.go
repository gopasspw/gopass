// +build !windows

package fsutil

import (
	"os"

	"golang.org/x/sys/unix"
)

// Tempdir returns a temporary directory suiteable for sensitive data. It tries
// /dev/shm but if this isn't working it will return an empty string. Using
// this with ioutil.Tempdir will ensure that we're getting the "best" tempdir.
func Tempdir() string {
	shmDir := "/dev/shm"
	if fi, err := os.Stat(shmDir); err == nil {
		if fi.IsDir() {
			if unix.Access(shmDir, unix.W_OK) == nil {
				return shmDir
			}
		}
	}
	return ""
}
