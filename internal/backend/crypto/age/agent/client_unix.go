//go:build !windows

package agent

import (
	"fmt"
	"os"
	"syscall"
)

func (c *Client) checkSocketSecurity() error {
	info, err := os.Stat(c.socketPath)
	if err != nil {
		return fmt.Errorf("failed to stat socket: %w", err)
	}

	// Check socket permissions.
	if info.Mode()&os.ModePerm != 0o600 {
		return fmt.Errorf("incorrect socket permissions: %v", info.Mode().Perm())
	}

	// Check socket ownership.
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return fmt.Errorf("failed to get socket system info")
	}

	if stat.Uid != uint32(os.Getuid()) {
		return fmt.Errorf("socket owned by wrong user: %d", stat.Uid)
	}

	return nil
}
