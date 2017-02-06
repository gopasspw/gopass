// +build windows

package fsutil

// Tempdir returns a temporary directory suiteable for sensitive data. On
// Windows, just return empty string for ioutil.TempFile.
func Tempdir() string {
	return ""
}
