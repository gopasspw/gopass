//go:build !windows

package agecrypto

import (
	"context"
	"os"
	"os/exec"
	"syscall"
)

func startAgent(_ context.Context) error {
	cmd := exec.Command(os.Args[0], "age", "agent")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	return cmd.Start()
}
