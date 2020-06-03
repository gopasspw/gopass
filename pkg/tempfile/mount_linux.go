// +build linux

package tempfile

import (
	"context"
	"os"

	"golang.org/x/sys/unix"
)

var shmDir = "/dev/shm"

// tempdir returns a temporary directory suiteable for sensitive data. It tries
// /dev/shm but if this isn't working it will return an empty string. Using
// this with ioutil.Tempdir will ensure that we're getting the "best" tempdir.
func tempdirBase() string {
	if fi, err := os.Stat(shmDir); err == nil {
		if fi.IsDir() {
			if unix.Access(shmDir, unix.W_OK) == nil {
				return shmDir
			}
		}
	}
	return ""
}

func (t *File) mount(context.Context) error {
	_ = t.dev // to trick megacheck
	return nil
}

func (t *File) unmount(context.Context) error {
	return nil
}
