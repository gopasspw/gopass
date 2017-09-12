// +build !linux,!darwin

package fsutil

import "context"

// tempdir returns a temporary directory suiteable for sensitive data. On
// Windows, just return empty string for ioutil.TempFile.
func tempdirBase() string {
	return ""
}

func (t *tempfile) mount(context.Context) error {
	_ = t.dev // to trick megacheck
	return nil
}

func (t *tempfile) unmount(context.Context) error {
	return nil
}
