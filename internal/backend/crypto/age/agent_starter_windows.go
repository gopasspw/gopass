//go:build windows

package agecrypto

import (
	"context"
	"os"
	"os/exec"
)

func startAgent(ctx context.Context) error {
	cmd := exec.Command(os.Args[0], "age", "agent")
	return cmd.Start()
}
