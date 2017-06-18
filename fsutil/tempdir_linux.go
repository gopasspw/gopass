// +build linux

package fsutil

import (
	"os"

	"golang.org/x/sys/unix"
)

// tempdir returns a temporary directory suiteable for sensitive data. It tries
// /dev/shm but if this isn't working it will return an empty string. Using
// this with ioutil.Tempdir will ensure that we're getting the "best" tempdir.
func tempdirBase() string {
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

func (t *tempfile) mount() error {
	_ = t.dev // to trick megacheck
	return nil
}

func (t *tempfile) unmount() error {
	return nil
}
