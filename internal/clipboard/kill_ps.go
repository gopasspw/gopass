// +build darwin freebsd,amd64 linux solaris windows

package clipboard

import (
	ps "github.com/mitchellh/go-ps"
)

// killPrecedessors will kill any previous "gopass unclip" invocations to avoid
// erasing the clipboard prematurely in case the the same content is copied to
// the clipboard repeatedly
func killPrecedessors() error {
	procs, err := ps.Processes()
	if err != nil {
		return err
	}
	for _, proc := range procs {
		walkFn(proc.Pid(), killProc)
	}
	return nil
}
