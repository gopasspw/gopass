// +build !darwin,!windows

package pinentry

// GetBinary returns the binary name
func GetBinary() string {
	return "pinentry"
}
