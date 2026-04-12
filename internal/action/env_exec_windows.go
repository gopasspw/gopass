//go:build windows

package action

import "fmt"

// execReplace is not supported on Windows because syscall.Exec is not
// available on that platform.
func execReplace(args []string, env []string) error {
	return fmt.Errorf("--exec is not supported on Windows")
}
