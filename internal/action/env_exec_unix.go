//go:build !windows

package action

import (
	"fmt"
	"os/exec"
	"syscall"
)

// execReplace replaces the current gopass process with the given command using
// syscall.Exec. Unlike exec.Command, this means gopass itself disappears from
// the process table entirely: the subprocess becomes the process, so its
// /proc/<pid>/environ no longer shows a gopass parent holding secrets.
// env must be the complete desired environment (typically os.Environ() plus
// the secret key/value pairs).
func execReplace(args []string, env []string) error {
	path, err := exec.LookPath(args[0])
	if err != nil {
		return fmt.Errorf("command %q not found: %w", args[0], err)
	}

	return syscall.Exec(path, args, env)
}
