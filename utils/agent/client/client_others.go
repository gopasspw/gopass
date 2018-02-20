// +build !windows

package client

import (
	"os"
	"os/exec"
	"syscall"
)

func (c *Client) startAgent() error {
	path, err := os.Executable()
	if err != nil {
		return err
	}

	cmd := exec.Command(path, "agent")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	return cmd.Start()
}
