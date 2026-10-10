package age

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/gopasspw/gopass/internal/backend/crypto/age/agent"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/termio"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

// shortTempDir returns a scratch directory whose path stays well under
// the unix socket path limit even on hosts with a long TMPDIR (macOS
// /var/folders/...): darwin caps sun_path at 104 bytes, linux at 108.
// GOPASS_HOMEDIR isolation is used instead of XDG_RUNTIME_DIR because
// windows resolves the runtime dir from LOCALAPPDATA and would ignore the
// XDG variable.
func shortTempDir(t *testing.T) string {
	t.Helper()

	base := "/tmp"
	if runtime.GOOS == "windows" {
		base = os.TempDir()
	}
	// os.MkdirTemp rather than t.TempDir: the latter follows TMPDIR,
	// which on macOS hosts exceeds the darwin sun_path limit (104 bytes)
	// once the socket path is appended.
	//nolint:usetesting
	dir, err := os.MkdirTemp(base, "gopass-agent-")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})

	return dir
}

// agentSocketPath isolates the client's socket resolution into a scratch
// directory and returns the exact path the client will dial, guarding the
// platform's unix socket path limit. Simulated agents must listen on this
// path: the runtime-dir convention differs per platform.
func agentSocketPath(t *testing.T) string {
	t.Helper()

	t.Setenv("GOPASS_HOMEDIR", shortTempDir(t))
	p := agent.NewClient().SocketPath()
	if len(p) > 103 {
		t.Skipf("unix socket path too long (%d bytes)", len(p))
	}

	return p
}

// startAgent starts an in-process age agent at the isolated socket path
// and waits until it answers pings (mirroring startTestAgent in the agent
// package tests).
func startAgent(t *testing.T) {
	t.Helper()

	agentSocketPath(t) // isolate + guard; the agent serves at this path

	ctx := t.Context()
	ctx = termio.WithPassPromptFunc(ctx, func(ctx context.Context, prompt string) (string, error) {
		return "test", nil
	})

	a, err := agent.New()
	require.NoError(t, err)
	go func() {
		_ = a.Run(ctx)
	}()
	t.Cleanup(func() {
		a.Shutdown(ctx)
	})

	c := agent.NewClient()
	require.Eventually(t, func() bool {
		return c.Ping() == nil
	}, 5*time.Second, 50*time.Millisecond, "age agent did not become ready")
}

// TestAgentStatus covers the status branches: not running (error return),
// running with capability negotiation, running against a legacy (pre-hello)
// agent, and an agent that speaks hello without a version= token. Each
// branch resolves its own isolated socket path so the test never touches -
// or kills - a real user agent.
func TestAgentStatus(t *testing.T) {
	ctx := t.Context()
	ctx = termio.WithPassPromptFunc(ctx, func(ctx context.Context, prompt string) (string, error) {
		return "test", nil
	})
	l := loader{}
	cmd := &cli.Command{}

	orig := out.Stdout
	defer func() {
		out.Stdout = orig
	}()
	capture := func() *bytes.Buffer {
		buf := &bytes.Buffer{}
		out.Stdout = buf

		return buf
	}

	// branch 1: agent not running -> error return, message printed
	agentSocketPath(t)

	buf := capture()
	err := l.agentStatus(ctx, cmd)
	require.Error(t, err)
	require.Contains(t, err.Error(), "agent not running")
	require.Contains(t, buf.String(), "Age agent is not running")

	// branch 2: agent running and hello-capable -> no error, capabilities
	// displayed
	startAgent(t)

	// arm an auto-lock timer so the countdown is reported
	require.NoError(t, agent.NewClient().SetTimeout(3600))

	buf = capture()
	require.NoError(t, l.agentStatus(ctx, cmd))
	require.Contains(t, buf.String(), "Age agent is running")
	require.Contains(t, buf.String(), "(auto-locks in ")
	require.Contains(t, buf.String(), "(agent version ")
	require.Contains(t, buf.String(), "Capabilities: ")
	require.NoError(t, agent.NewClient().Quit())

	// branch 3: legacy agent (pre-hello) that still serves status but
	// answers ERR to hello -> no error, legacy note printed
	legacySock := agentSocketPath(t)
	require.NoError(t, os.MkdirAll(filepath.Dir(legacySock), 0o700))
	ln, err := net.Listen("unix", legacySock)
	require.NoError(t, err)
	defer func() {
		_ = ln.Close()
	}()
	require.NoError(t, os.Chmod(legacySock, 0o600))

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func() {
				defer func() {
					_ = conn.Close()
				}()
				sc := bufio.NewScanner(conn)
				for sc.Scan() {
					if sc.Text() == "status" {
						fmt.Fprintln(conn, "OK")

						continue
					}
					fmt.Fprintln(conn, "ERR unknown command")
				}
			}()
		}
	}()

	buf = capture()
	require.NoError(t, l.agentStatus(ctx, cmd))
	require.Contains(t, buf.String(), "Age agent is running")
	require.Contains(t, buf.String(), "(legacy agent, no capability negotiation)")

	// branch 4: agent answers hello but advertises no version= token (e.g.
	// a third-party implementation) -> capabilities shown, no version line
	capsSock := agentSocketPath(t)
	require.NoError(t, os.MkdirAll(filepath.Dir(capsSock), 0o700))
	ln2, err := net.Listen("unix", capsSock)
	require.NoError(t, err)
	defer func() {
		_ = ln2.Close()
	}()
	require.NoError(t, os.Chmod(capsSock, 0o600))

	go func() {
		for {
			conn, err := ln2.Accept()
			if err != nil {
				return
			}
			go func() {
				defer func() {
					_ = conn.Close()
				}()
				sc := bufio.NewScanner(conn)
				for sc.Scan() {
					switch {
					case sc.Text() == "status":
						fmt.Fprintln(conn, "OK")
					case strings.HasPrefix(sc.Text(), "hello"):
						fmt.Fprintln(conn, "OK ping")
					default:
						fmt.Fprintln(conn, "ERR unknown command")
					}
				}
			}()
		}
	}()

	buf = capture()
	require.NoError(t, l.agentStatus(ctx, cmd))
	require.Contains(t, buf.String(), "Age agent is running")
	require.Contains(t, buf.String(), "Capabilities: ping")
	require.NotContains(t, buf.String(), "(agent version")
}
