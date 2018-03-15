// +build windows

package client

import (
	"context"
	"os"
	"os/exec"
	"syscall"
)

const (
	// CREATE_NEW_PROCESS_GROUP is like Setpgid on UNIX
	// https://msdn.microsoft.com/en-us/library/windows/desktop/ms684863(v=vs.85).aspx
	CREATE_NEW_PROCESS_GROUP = 0x00000200
	// DETACHED_PROCESS does not inherit the parent console
	DETACHED_PROCESS = 0x00000008
)

func (c *Client) startAgent(ctx context.Context) error {
	path, err := os.Executable()
	if err != nil {
		return err
	}

	cmd := exec.Command(path, "agent")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: CREATE_NEW_PROCESS_GROUP | DETACHED_PROCESS,
	}
	return cmd.Start()
}
