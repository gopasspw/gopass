//go:build !openbsd

// Package protect provides an interface to the pledge syscall.
// It is used to limit the system calls a process can make.
// This is used to limit the attack surface of the process.
// The pledge syscall is only available on OpenBSD.
// It is not available on other systems.
// This package is a no-op on other systems.
package protect

// ProtectEnabled lets us know if we have protection or not.
// It is false on all systems except OpenBSD.
var ProtectEnabled = false

// Pledge on any other system than OpenBSD doesn't do anything.
func Pledge(s string) error {
	return nil
}
