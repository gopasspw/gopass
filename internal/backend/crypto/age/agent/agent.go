// Package agent implements the gopass age-agent.
package agent

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"maps"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"filippo.io/age"
	"github.com/gopasspw/gopass/internal/buildinfo"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/appdir"
	"github.com/gopasspw/gopass/pkg/debug"
)

const (
	socketName = "gopass-age-agent.sock"

	// maxLineSize is the maximum line length accepted on a connection. The
	// decrypt command carries the base64-encoded ciphertext on a single
	// line, so bufio's default 64 KiB scanner limit must be raised to cover
	// large secrets (see issue #3508).
	maxLineSize = 1 << 24 // 16 MiB
)

// agentHandler serves one command on one connection. It returns false when
// the connection must be closed after the command (only quit does).
type agentHandler func(a *Agent, ctx context.Context, conn net.Conn, args []string) bool

// agentHandlers maps the advertised wire commands to their
// implementations. The hello capability list is DERIVED from these keys,
// so the advertised set can never drift from what is dispatched.
var agentHandlers = map[string]agentHandler{
	"decrypt":     (*Agent).handleDecrypt,
	"identities":  (*Agent).handleIdentities,
	"lock":        (*Agent).handleLock,
	"ping":        (*Agent).handlePing,
	"quit":        (*Agent).handleQuit,
	"set-timeout": (*Agent).handleSetTimeout,
	"status":      (*Agent).handleStatus,
	"unlock":      (*Agent).handleUnlock,
}

// capabilityTokens is the sorted command-token list of the hello response,
// precomputed once: agentHandlers is immutable after initialization, so
// only the dynamic tokens (maxline, version) are appended per request.
var capabilityTokens = slices.Sorted(maps.Keys(agentHandlers))

// commandHandlers is the full dispatch table: every advertised command
// plus hello, which is dispatched like any other command but deliberately
// not advertised — it is the negotiation channel, not a capability.
var commandHandlers = func() map[string]agentHandler {
	m := maps.Clone(agentHandlers)
	m["hello"] = (*Agent).handleHello

	return m
}()

// Agent is a gopass age agent.
type Agent struct {
	socketPath string
	listener   net.Listener

	mux        sync.Mutex
	identities []age.Identity
	locked     bool
	timer      *time.Timer
	timeout    time.Duration
}

// New creates a new agent.
func New() (*Agent, error) {
	socketDir := appdir.UserRuntime()
	if err := os.MkdirAll(socketDir, 0o700); err != nil {
		return nil, fmt.Errorf("failed to create socket directory: %w", err)
	}

	socketPath := filepath.Join(socketDir, socketName)

	return &Agent{
		socketPath: socketPath,
		locked:     false,
		timeout:    0,
	}, nil
}

// Run starts the agent.
func (a *Agent) Run(ctx context.Context) error {
	// listen on the socket
	l, err := net.Listen("unix", a.socketPath)
	if err != nil {
		return fmt.Errorf("failed to listen on socket: %w", err)
	}
	if err := os.Chmod(a.socketPath, 0o600); err != nil {
		return fmt.Errorf("failed to set socket permissions: %w", err)
	}
	a.listener = l
	defer func() {
		_ = a.listener.Close()
	}()

	debug.Log("agent listening on %s", a.socketPath)

	// handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		debug.Log("received signal %s, shutting down", sig)
		a.Shutdown(ctx)
	}()

	// accept connections
	for {
		conn, err := a.listener.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				return nil
			}
			debug.Log("failed to accept connection: %s", err)

			continue
		}
		go a.handleConnection(ctx, conn)
	}
}

// Shutdown stops the agent.
func (a *Agent) Shutdown(ctx context.Context) {
	if a.listener != nil {
		_ = a.listener.Close()
	}
	if err := os.Remove(a.socketPath); err != nil {
		debug.Log("failed to remove socket file: %s", err)
	}

	debug.Log("agent shut down")
}

// capabilities renders the space-separated capability tokens returned by
// the hello command. The command tokens are precomputed in
// capabilityTokens (derived from agentHandlers, so the advertised set
// cannot drift from dispatch, and sorted for a deterministic wire format).
// The version token is diagnostic only and must never be used to gate
// behaviour.
func capabilities() string {
	toks := make([]string, 0, len(capabilityTokens)+2)
	toks = append(toks, capabilityTokens...)
	toks = append(toks, "maxline="+strconv.Itoa(maxLineSize))
	if v, ok := buildinfo.ModuleVersion(); ok {
		toks = append(toks, "version="+v)
	} else {
		toks = append(toks, "version=unknown")
	}

	return strings.Join(toks, " ")
}

func (a *Agent) handleConnection(ctx context.Context, conn net.Conn) {
	defer func() {
		_ = conn.Close()
	}()

	scanner := bufio.NewScanner(conn)
	// no buffer is preallocated: the scanner starts at bufio's small
	// initial size and only doubles as the longest received line requires,
	// up to maxLineSize.
	scanner.Buffer(nil, maxLineSize)
	for scanner.Scan() {
		line := scanner.Text()

		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		cmd := parts[0]
		args := parts[1:]

		if handler, ok := commandHandlers[cmd]; ok {
			if !handler(a, ctx, conn, args) {
				return
			}

			continue
		}

		// The first token of an unknown command may itself be a
		// mis-framed payload, e.g. a bare AGE-SECRET-KEY-1... line from
		// a newline-separated identities send (see identitiesToString in
		// the age package). Never log it verbatim: the 12-char prefix
		// (mirroring wrappedIdentity.SafeStr) reveals at most the
		// constant key-type prefix, never key material.
		debug.Log("received: unknown command %.12s... (%d bytes)", cmd, len(line))

		fmt.Fprintln(conn, "ERR unknown command")
	}
	if err := scanner.Err(); err != nil {
		debug.Log("agent connection scan error: %s", err)
	}
}

func (a *Agent) handlePing(_ context.Context, conn net.Conn, _ []string) bool {
	debug.Log("received: ping")

	fmt.Fprintln(conn, "OK")

	return true
}

func (a *Agent) handleStatus(_ context.Context, conn net.Conn, _ []string) bool {
	debug.Log("received: status")

	a.mux.Lock()
	locked := a.locked
	a.mux.Unlock()
	if locked {
		fmt.Fprintln(conn, "OK locked")
	} else {
		fmt.Fprintln(conn, "OK")
	}

	return true
}

func (a *Agent) handleIdentities(_ context.Context, conn net.Conn, args []string) bool {
	// The arguments contain private key material. Each identity is
	// logged and hidden unless GOPASS_DEBUG_LOG_SECRETS is set
	debug.Log("received: identities [%d ids]", len(args))
	for _, id := range args {
		debug.Log("received: identity: %s", out.Secret(id))
	}

	if len(args) < 1 {
		fmt.Fprintln(conn, "ERR missing identities")

		return true
	}
	ids, err := parseIdentities(strings.NewReader(strings.Join(args, "\n")))
	if err != nil {
		fmt.Fprintln(conn, "ERR failed to parse identities: "+err.Error())

		return true
	}
	a.mux.Lock()
	a.identities = ids
	a.mux.Unlock()
	debug.Log("loaded %d identities", len(ids))
	fmt.Fprintln(conn, "OK")

	return true
}

func (a *Agent) handleDecrypt(_ context.Context, conn net.Conn, args []string) bool {
	// the argument is base64 ciphertext: log only its encoded size,
	// the relevant diagnostic for the scanner line-size limit
	// (see issue #3508)
	if len(args) == 1 {
		debug.Log("received: decrypt (%d bytes)", len(args[0]))
	} else {
		debug.Log("received: decrypt")
	}

	if len(args) != 1 {
		fmt.Fprintln(conn, "ERR missing ciphertext")

		return true
	}
	ciphertext, err := base64.StdEncoding.DecodeString(args[0])
	if err != nil {
		fmt.Fprintln(conn, "ERR failed to decode ciphertext: "+err.Error())

		return true
	}
	plaintext, err := a.decrypt(ciphertext)
	if err != nil {
		if err.Error() == "agent is locked" {
			fmt.Fprintln(conn, "ERR agent is locked")
		} else {
			fmt.Fprintln(conn, "ERR failed to decrypt: "+err.Error())
		}

		return true
	}
	fmt.Fprintln(conn, "OK "+base64.StdEncoding.EncodeToString(plaintext))

	return true
}

func (a *Agent) handleLock(_ context.Context, conn net.Conn, _ []string) bool {
	debug.Log("received: lock")

	// clear all identities from memory
	a.mux.Lock()
	a.identities = nil
	a.locked = true
	a.mux.Unlock()

	debug.Log("cleared identities from memory and locked agent")
	fmt.Fprintln(conn, "OK")

	return true
}

func (a *Agent) handleUnlock(_ context.Context, conn net.Conn, _ []string) bool {
	debug.Log("received: unlock")

	a.mux.Lock()
	a.locked = false
	a.mux.Unlock()

	debug.Log("unlocked agent")
	fmt.Fprintln(conn, "OK")

	return true
}

func (a *Agent) handleSetTimeout(_ context.Context, conn net.Conn, args []string) bool {
	// the timeout value is not sensitive
	if len(args) > 0 {
		debug.Log("received: set-timeout %s", args[0])
	} else {
		debug.Log("received: set-timeout")
	}

	if len(args) != 1 {
		fmt.Fprintln(conn, "ERR missing timeout")

		return true
	}
	timeout, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintln(conn, "ERR failed to parse timeout: "+err.Error())

		return true
	}
	a.setTimeout(time.Duration(timeout) * time.Second)
	fmt.Fprintln(conn, "OK")

	return true
}

// handleQuit replies OK and shuts the agent down. It returns false so the
// serving connection is closed immediately after.
func (a *Agent) handleQuit(ctx context.Context, conn net.Conn, _ []string) bool {
	debug.Log("received: quit")

	fmt.Fprintln(conn, "OK")
	go a.Shutdown(ctx)

	return false
}

// handleHello is a stateless, voluntary capability exchange (issue #3624):
//
//	client> hello [gopass/<version>]   (argument ignored beyond diagnostics)
//	agent<  OK <command tokens...> maxline=<n> version=<v>
//
// No per-connection or global state is kept; receiving or not receiving
// hello must never change any other command's behaviour. Clients detect a
// pre-hello (legacy) agent by hello yielding any ERR response, without
// matching the error text.
func (a *Agent) handleHello(_ context.Context, conn net.Conn, args []string) bool {
	// the client version argument is not sensitive (a self-reported
	// gopass/x.y.z string) and is diagnostics-only
	if len(args) > 0 {
		debug.Log("received: hello %s", args[0])
	} else {
		debug.Log("received: hello")
	}

	fmt.Fprintln(conn, "OK "+capabilities())

	return true
}

func (a *Agent) setTimeout(timeout time.Duration) {
	a.mux.Lock()
	defer a.mux.Unlock()

	a.timeout = timeout
	if a.timer != nil {
		a.timer.Stop()
	}
	if a.timeout > 0 {
		a.timer = time.AfterFunc(a.timeout, func() {
			a.lock()
		})
	}
}

func (a *Agent) lock() {
	a.mux.Lock()
	defer a.mux.Unlock()

	a.identities = nil
	a.locked = true
	if a.timer != nil {
		a.timer.Stop()
	}
	debug.Log("cleared identities from memory and locked agent")
}

func (a *Agent) decrypt(ciphertext []byte) ([]byte, error) {
	a.mux.Lock()
	defer a.mux.Unlock()
	if a.locked {
		return nil, fmt.Errorf("agent is locked")
	}
	if a.timer != nil {
		a.timer.Reset(a.timeout)
	}
	out := &bytes.Buffer{}
	f := bytes.NewReader(ciphertext)
	r, err := age.Decrypt(f, a.identities...)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	if _, err := io.Copy(out, r); err != nil {
		return nil, fmt.Errorf("failed to write plaintext to buffer: %w", err)
	}

	return out.Bytes(), nil
}
