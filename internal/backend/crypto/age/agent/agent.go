// Package agent implements the gopass age-agent.
package agent

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"filippo.io/age"
	"filippo.io/age/agessh"
	"github.com/gopasspw/gopass/internal/backend/crypto/age/identityorder"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/appdir"
	"github.com/gopasspw/gopass/pkg/debug"
)

const (
	socketName = "gopass-age-agent.sock"
)

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

func (a *Agent) handleConnection(ctx context.Context, conn net.Conn) {
	defer func() {
		_ = conn.Close()
	}()

	scanner := bufio.NewScanner(conn)
	// the decrypt command carries the base64-encoded ciphertext on a single
	// line, so we need to raise the default 64 KiB token limit to handle
	// larger secrets (see issue #3508). No buffer is preallocated: the
	// scanner starts at bufio's small initial size and only doubles as the
	// longest received line requires, up to this limit. The 16 MiB maximum
	// matches the privateKeySizeLimit used when parsing identities.
	scanner.Buffer(nil, 1<<24) // 16 MiB max line size
	for scanner.Scan() {
		line := scanner.Text()

		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		cmd := parts[0]
		args := parts[1:]

		switch cmd {
		case "ping":
			debug.Log("received: ping")

			fmt.Fprintln(conn, "OK")
		case "status":
			debug.Log("received: status")

			a.mux.Lock()
			locked := a.locked
			a.mux.Unlock()
			if locked {
				fmt.Fprintln(conn, "OK locked")
			} else {
				fmt.Fprintln(conn, "OK")
			}
		case "identities":
			// The arguments contain private key material. Each identity is
			// logged and hidden unless GOPASS_DEBUG_LOG_SECRETS is set
			debug.Log("received: identities [%d ids]", len(args))
			for _, id := range args {
				debug.Log("received: identity: %s", out.Secret(id))
			}

			if len(args) < 1 {
				fmt.Fprintln(conn, "ERR missing identities")

				continue
			}
			ids, err := parseIdentities(strings.NewReader(strings.Join(args, "\n")))
			if err != nil {
				fmt.Fprintln(conn, "ERR failed to parse identities: "+err.Error())

				continue
			}
			a.mux.Lock()
			a.identities = ids
			a.mux.Unlock()
			debug.Log("loaded %d identities", len(ids))
			fmt.Fprintln(conn, "OK")
		case "ssh-identity":
			fmt.Fprintln(conn, a.addSSHIdentityResponse(args))
		case "decrypt":
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

				continue
			}
			ciphertext, err := base64.StdEncoding.DecodeString(args[0])
			if err != nil {
				fmt.Fprintln(conn, "ERR failed to decode ciphertext: "+err.Error())

				continue
			}
			plaintext, err := a.decrypt(ciphertext)
			if err != nil {
				if err.Error() == "agent is locked" {
					fmt.Fprintln(conn, "ERR agent is locked")
				} else {
					fmt.Fprintln(conn, "ERR failed to decrypt: "+err.Error())
				}

				continue
			}
			fmt.Fprintln(conn, "OK "+base64.StdEncoding.EncodeToString(plaintext))
		case "lock":
			debug.Log("received: lock")

			// clear all identities from memory
			a.mux.Lock()
			a.identities = nil
			a.locked = true
			a.mux.Unlock()

			debug.Log("cleared identities from memory and locked agent")
			fmt.Fprintln(conn, "OK")
		case "unlock":
			debug.Log("received: unlock")

			a.mux.Lock()
			a.locked = false
			a.mux.Unlock()

			debug.Log("unlocked agent")
			fmt.Fprintln(conn, "OK")
		case "set-timeout":
			fmt.Fprintln(conn, a.setTimeoutResponse(args))
		case "quit":
			debug.Log("received: quit")

			fmt.Fprintln(conn, "OK")
			go a.Shutdown(ctx)

			return
		default:
			// The first token of an unknown command may itself be a
			// mis-framed payload, e.g. a bare AGE-SECRET-KEY-1... line from
			// a newline-separated identities send (see identitiesToString in
			// the age package). Never log it verbatim: the 12-char prefix
			// (mirroring wrappedIdentity.SafeStr) reveals at most the
			// constant key-type prefix, never key material.
			debug.Log("received: unknown command %.12s... (%d bytes)", cmd, len(line))

			fmt.Fprintln(conn, "ERR unknown command")
		}
	}
	if err := scanner.Err(); err != nil {
		debug.Log("agent connection scan error: %s", err)
	}
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
	r, err := age.Decrypt(f, identityorder.Preserve(a.identities)...)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	if _, err := io.Copy(out, r); err != nil {
		return nil, fmt.Errorf("failed to write plaintext to buffer: %w", err)
	}

	return out.Bytes(), nil
}
