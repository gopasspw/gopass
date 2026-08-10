//go:build !windows

package ageagentlauncher

import (
	"context"
	"os"
	"os/exec"
	"testing"

	"github.com/gopasspw/gopass/internal/backend/crypto/age"
)

// TestLaunchStampsSpawnGuard proves launch() stamps age.SpawnGuardEnv on the
// spawned child. It overrides execCommand to re-exec the test binary as
// TestLaunchHelper, which exits non-zero unless it inherited SpawnGuardEnv=1.
// This locks the defense-in-depth layer: if the cmd.Env line in launch() is
// ever dropped, the helper fails and this test catches it.
func TestLaunchStampsSpawnGuard(t *testing.T) {
	// Carried into the child via launch()'s os.Environ(); tells the helper to run.
	t.Setenv("GP_AGENTLAUNCH_TEST_HELPER", "1")

	var spawned *exec.Cmd
	orig := execCommand
	execCommand = func(name string, args ...string) *exec.Cmd {
		spawned = exec.Command(os.Args[0], append([]string{"-test.run=TestLaunchHelper", "--", name}, args...)...)

		return spawned
	}
	t.Cleanup(func() { execCommand = orig })

	if err := launch(context.Background()); err != nil {
		t.Fatalf("launch returned error: %v", err)
	}
	if spawned == nil {
		t.Fatal("launch did not build a command")
	}
	if err := spawned.Wait(); err != nil {
		t.Fatalf("spawned child did not see %s=1 (launch must stamp it): %v", age.SpawnGuardEnv, err)
	}
}

// TestLaunchHelper is the child side of TestLaunchStampsSpawnGuard. It runs only
// when re-exec'd with GP_AGENTLAUNCH_TEST_HELPER=1, and fails unless it inherited
// SpawnGuardEnv=1 from launch()'s cmd.Env. Skips during normal test runs.
func TestLaunchHelper(t *testing.T) {
	if os.Getenv("GP_AGENTLAUNCH_TEST_HELPER") != "1" {
		t.Skip()
	}
	if os.Getenv(age.SpawnGuardEnv) != "1" {
		t.Fatalf("%s not set on spawned child", age.SpawnGuardEnv)
	}
}
