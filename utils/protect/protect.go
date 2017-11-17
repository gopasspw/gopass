// +build !openbsd

package protect

// ProtectEnabled lets us know if we have protection or not
var ProtectEnabled = false

// Pledge on any other system than OpenBSD doesn't do anything
func Pledge(s string) {
	return
}
