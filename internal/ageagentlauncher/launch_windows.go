//go:build windows

package ageagentlauncher

import (
	"context"
	"os"

	"github.com/gopasspw/gopass/internal/backend/crypto/age"
)

func launch(_ context.Context) error {
	cmd := execCommand(os.Args[0], "age", "agent", "start")
	cmd.Env = append(os.Environ(), age.SpawnGuardEnv+"=1")
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	return cmd.Start()
}
