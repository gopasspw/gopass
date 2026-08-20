// Package ageagentlauncher wires the standalone gopass CLI's age-agent launcher
// into the age backend.
//
// Only the standalone gopass binary owns the `age agent start` subcommand, so
// the os.Args[0] re-exec that starts the agent lives here (CLI side), not in
// the age backend. The backend must stay embeddable as a library: a consumer
// that does not register a launcher gets graceful degradation (no autostart,
// per-operation password prompt) instead of a fork bomb. See
// internal/backend/crypto/age.WithAgentLauncher.
package ageagentlauncher

import (
	"context"
	"os/exec"

	"github.com/gopasspw/gopass/internal/backend/crypto/age"
)

// execCommand is indirection over os/exec.Command so tests can substitute a
// re-exec of the test binary and assert on the spawned child's environment
// without spawning a real agent.
var execCommand = exec.Command

// Register installs the CLI's age-agent launcher into ctx so the age backend
// can auto-start the agent when needed. Call once during CLI startup (e.g. from
// main.initContext); it covers every age.New call site because the top-level
// ctx propagates to every command action.
func Register(ctx context.Context) context.Context {
	return age.WithAgentLauncher(ctx, launch)
}
