// +build !linux,!darwin

package tempfile

import "context"

var shmDir = ""

// tempdir returns a temporary directory suiteable for sensitive data. On
// Windows, just return empty string for ioutil.TempFile.
func tempdirBase() string {
	return ""
}

func (t *File) mount(context.Context) error {
	_ = t.dev // to trick megacheck
	return nil
}

func (t *File) unmount(context.Context) error {
	return nil
}
