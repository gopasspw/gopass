// +build !linux,!darwin

package fsutil

// tempdir returns a temporary directory suiteable for sensitive data. On
// Windows, just return empty string for ioutil.TempFile.
func tempdirBase() string {
	return ""
}

func (t *tempfile) mount() error {
	_ = t.dev // to trick megacheck
	return nil
}

func (t *tempfile) unmount() error {
	return nil
}
