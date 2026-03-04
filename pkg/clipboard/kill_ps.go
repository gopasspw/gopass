//go:build darwin || (freebsd && amd64) || linux || solaris || windows || (freebsd && arm) || (freebsd && arm64)

package clipboard

import (
	"fmt"

	ps "github.com/mitchellh/go-ps"
)

// killPrecedessors will kill any previous "gopass unclip" invocations to avoid
// erasing the clipboard prematurely in case the the same content is copied to
// the clipboard repeatedly.
func killPrecedessors() error {
	procs, err := ps.Processes()
	if err != nil {
		return fmt.Errorf("failed to list processes: %w", err)
	}

	for _, proc := range procs {
		walkFn(proc.Pid(), killProc)
	}

	return nil
}
