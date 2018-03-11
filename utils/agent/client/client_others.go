// +build !windows

package client

import (
	"context"
	"os"
	"os/exec"
	"syscall"

	"github.com/justwatchcom/gopass/utils/out"
	"github.com/pkg/errors"
)

func (c *Client) startAgent(ctx context.Context) error {
	path, err := os.Executable()
	if err != nil {
		return errors.Wrapf(err, "unable to determine executable: %s", err)
	}

	out.Debug(ctx, "Starting agent ...")
	cmd := exec.Command(path, "agent")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	return cmd.Start()
}
