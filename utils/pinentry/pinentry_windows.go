// +build windows

package pinentry

// GetBinary always returns pinentry.exe
func GetBinary() string {
	return "pinentry.exe"
}
