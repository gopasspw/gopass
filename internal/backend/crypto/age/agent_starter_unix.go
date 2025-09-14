//go:build !windows

package age

import (
	"context"
	"os"
	"os/exec"
	"syscall"
)

func startAgent(ctx context.Context) error {
	cmd := exec.Command(os.Args[0], "age", "agent")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	return cmd.Start()
}
