//go:build !windows

package age

import (
	"context"
	"os"
	"os/exec"
	"syscall"
)

func startAgent(_ context.Context) error {
	cmd := exec.Command(os.Args[0], "age", "agent", "start")
	cmd.Env = os.Environ()
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	return cmd.Start()
}
