// +build darwin

package pinentry

// GetBinary always returns pinentry-mac
func GetBinary() string {
	return "pinentry-mac"
}
