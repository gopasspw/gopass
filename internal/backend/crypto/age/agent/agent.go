// Package agent implements the gopass age-agent.
package agent

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
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
	"filippo.io/age/agessh"
	"github.com/gopasspw/gopass/internal/backend/crypto/age/identityorder"
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
	// large secrets (see issue #3508). It matches the privateKeySizeLimit
	// used when parsing identities.
	maxLineSize = 1 << 24 // 16 MiB
)

var ErrAgentLocked = fmt.Errorf("agent is locked")

// agentHandler serves one command on one connection. It returns false when
// the connection must be closed after the command (only quit does).
type agentHandler func(ctx context.Context, conn net.Conn, args []string) bool

// Agent is a gopass age agent.
type Agent struct {
	socketPath string
	listener   net.Listener

	mux         sync.Mutex
	identities  []age.Identity
	source      string
	unlockToken string
	locked      bool
	stopped     bool
	timer       *time.Timer
	timeout     time.Duration
	generation  uint64
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
		_ = l.Close()

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
	a.mux.Lock()
	a.stopped = true
	a.clearLocked()
	a.mux.Unlock()
	if a.listener != nil {
		_ = a.listener.Close()
	}
	if err := os.Remove(a.socketPath); err != nil {
		debug.Log("failed to remove socket file: %s", err)
	}

	debug.Log("agent shut down")
}

// commandHandlers returns the agent's dispatch table: every wire command
// mapped to its bound method, hello included. Both dispatch and the
// advertised capability list derive from this table, so the two cannot
// drift apart. It is deliberately built per call rather than declared as
// a package-level variable: a package-level table containing hello, with
// capabilities() deriving from it, forms a Go initialization cycle.
func (a *Agent) commandHandlers() map[string]agentHandler {
	return map[string]agentHandler{
		"decrypt":      a.handleDecrypt,
		"hello":        a.handleHello,
		"identities":   a.handleIdentities,
		"lock":         a.handleLock,
		"ping":         a.handlePing,
		"quit":         a.handleQuit,
		"session":      a.handleSessionCommand,
		"set-timeout":  a.handleSetTimeout,
		"ssh-identity": a.handleSSHIdentity,
		"status":       a.handleStatus,
		"unlock":       a.handleUnlock,
	}
}

// capabilities renders the space-separated capability tokens returned by
// the hello command: every dispatchable command except hello itself (the
// negotiation command advertises the others, not itself), plus maxline
// and the version token. The version token is diagnostic only and must
// never be used to gate behaviour.
func (a *Agent) capabilities() string {
	handlers := a.commandHandlers()
	toks := make([]string, 0, len(handlers))
	for cmd := range handlers {
		if cmd == "hello" {
			continue
		}
		toks = append(toks, cmd)
	}
	slices.Sort(toks)
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
	handlers := a.commandHandlers()
	for scanner.Scan() {
		line := scanner.Text()

		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		cmd := parts[0]
		args := parts[1:]

		if handler, ok := handlers[cmd]; ok {
			if !handler(ctx, conn, args) {
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
	a.loadLegacyIdentities(conn, args)

	return true
}

func (a *Agent) handleSSHIdentity(_ context.Context, conn net.Conn, args []string) bool {
	fmt.Fprintln(conn, a.addSSHIdentityResponse(args))

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
		if errors.Is(err, ErrAgentLocked) {
			fmt.Fprintln(conn, "ERR agent is locked")

			return true
		}

		fmt.Fprintln(conn, "ERR failed to decrypt: "+err.Error())

		return true
	}
	fmt.Fprintln(conn, "OK "+base64.StdEncoding.EncodeToString(plaintext))

	return true
}

func (a *Agent) handleLock(_ context.Context, conn net.Conn, _ []string) bool {
	debug.Log("received: lock")

	a.lock()
	fmt.Fprintln(conn, "OK")

	return true
}

func (a *Agent) handleUnlock(_ context.Context, conn net.Conn, _ []string) bool {
	debug.Log("received: unlock")

	if err := a.unlock(); err != nil {
		fmt.Fprintln(conn, "ERR "+err.Error())

		return true
	}

	debug.Log("unlocked agent")
	fmt.Fprintln(conn, "OK")

	return true
}

func (a *Agent) handleSetTimeout(_ context.Context, conn net.Conn, args []string) bool {
	fmt.Fprintln(conn, a.setTimeoutResponse(args))

	return true
}

// handleSessionCommand adapts the source-bound session protocol (see
// session.go) to the common dispatch table.
func (a *Agent) handleSessionCommand(_ context.Context, conn net.Conn, args []string) bool {
	a.handleSession(conn, args)

	return true
}

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

	fmt.Fprintln(conn, "OK "+a.capabilities())

	return true
}

func (a *Agent) addSSHIdentityResponse(args []string) string {
	if err := a.addSSHIdentity(args); err != nil {
		return "ERR " + err.Error()
	}

	return "OK"
}

func (a *Agent) addSSHIdentity(args []string) error {
	// The argument is an unencrypted OpenSSH private key. Log only its encoded
	// size and never the payload itself.
	if len(args) != 1 {
		debug.Log("received: ssh-identity")

		return fmt.Errorf("missing SSH identity")
	}
	debug.Log("received: ssh-identity (%d bytes)", len(args[0]))

	privateKey, err := base64.StdEncoding.DecodeString(args[0])
	if err != nil {
		return fmt.Errorf("failed to decode SSH identity: %w", err)
	}
	id, err := agessh.ParseIdentity(privateKey)
	clear(privateKey)
	if err != nil {
		return fmt.Errorf("failed to parse SSH identity: %w", err)
	}

	a.mux.Lock()
	if a.stopped {
		a.mux.Unlock()

		return fmt.Errorf("agent is stopped")
	}
	if a.source != "" || a.unlockToken != "" {
		a.mux.Unlock()

		return fmt.Errorf("cannot add external SSH identities to a keyring session")
	}
	a.identities = append(a.identities, id)
	a.mux.Unlock()
	debug.Log("loaded SSH identity")

	return nil
}

func (a *Agent) setTimeoutResponse(args []string) string {
	// The timeout value is not sensitive.
	if len(args) > 0 {
		debug.Log("received: set-timeout %s", args[0])
	} else {
		debug.Log("received: set-timeout")
	}

	if len(args) != 1 {
		return "ERR missing timeout"
	}
	timeout, err := strconv.Atoi(args[0])
	if err != nil {
		return "ERR failed to parse timeout: " + err.Error()
	}
	a.setTimeout(time.Duration(timeout) * time.Second)

	return "OK"
}

func (a *Agent) setTimeout(timeout time.Duration) {
	a.mux.Lock()
	defer a.mux.Unlock()

	if a.stopped {
		return
	}
	a.timeout = timeout
	a.startTimerLocked()
}

func (a *Agent) lock() {
	a.mux.Lock()
	defer a.mux.Unlock()

	a.clearLocked()
	debug.Log("cleared identities from memory and locked agent")
}

func (a *Agent) decrypt(ciphertext []byte) ([]byte, error) {
	a.mux.Lock()
	defer a.mux.Unlock()
	if a.source != "" {
		return nil, fmt.Errorf("session decryption requires source")
	}

	return a.decryptLocked(ciphertext)
}

// decryptLocked requires a.mux to be held.
func (a *Agent) decryptLocked(ciphertext []byte) ([]byte, error) {
	if a.stopped {
		return nil, fmt.Errorf("agent is stopped")
	}
	if a.locked {
		return nil, ErrAgentLocked
	}
	if a.timeout > 0 {
		a.startTimerLocked()
	}
	out := &bytes.Buffer{}
	f := bytes.NewReader(ciphertext)
	r, err := age.Decrypt(f, identityorder.Preserve(a.identities)...)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	if _, err := io.Copy(out, r); err != nil {
		return nil, fmt.Errorf("failed to read plaintext: %w", err)
	}

	return out.Bytes(), nil
}

func (a *Agent) loadLegacyIdentities(conn io.Writer, args []string) {
	// The arguments contain private key material. Each identity is
	// logged and hidden unless GOPASS_DEBUG_LOG_SECRETS is set
	debug.Log("received: identities [%d ids]", len(args))
	for _, id := range args {
		debug.Log("received: identity: %s", out.Secret(id))
	}

	if len(args) < 1 {
		fmt.Fprintln(conn, "ERR missing identities")

		return
	}
	ids, err := parseIdentities(strings.NewReader(strings.Join(args, "\n")))
	if err != nil {
		fmt.Fprintln(conn, "ERR failed to parse identities: "+err.Error())

		return
	}
	a.mux.Lock()
	if a.stopped {
		a.mux.Unlock()
		fmt.Fprintln(conn, "ERR agent is stopped")

		return
	}
	a.identities = ids
	a.source = ""
	a.startTimerLocked()
	a.mux.Unlock()
	debug.Log("loaded %d identities", len(ids))
	fmt.Fprintln(conn, "OK")
}

func (a *Agent) unlock() error {
	a.mux.Lock()
	defer a.mux.Unlock()
	if a.stopped {
		return fmt.Errorf("agent is stopped")
	}
	a.locked = false
	a.startTimerLocked()

	return nil
}
