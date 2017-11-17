// +build !openbsd

package protect

// Pledge on any other system than OpenBSD doesn't do anything
func Pledge(s string) {
	return
}
