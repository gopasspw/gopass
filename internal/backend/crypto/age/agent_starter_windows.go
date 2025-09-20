//go:build windows

package age

import (
	"context"
	"os"
	"os/exec"
)

func startAgent(_ context.Context) error {
	cmd := exec.Command(os.Args[0], "age", "agent", "start")
	cmd.Env = os.Environ()
	cmd.Stderr = os.Stderr
	return cmd.Start()
}
