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
	"strings"
	"syscall"
	"time"

	"filippo.io/age"
	"github.com/gopasspw/gopass/pkg/appdir"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/pinentry/cli"
	"github.com/twpayne/go-pinentry/v4"
)

const (
	socketName = "gopass-age-agent.sock"
)

// Agent is a gopass age agent.
type Agent struct {
	socketPath string
	listener   net.Listener
	cache      *InMemTTL[string, string]
	identities []age.Identity
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
		cache:      NewInMemTTL[string, string](time.Minute*5, time.Hour),
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
}

func (a *Agent) handleConnection(ctx context.Context, conn net.Conn) {
	defer func() {
		_ = conn.Close()
	}()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		debug.Log("received: %s", line)

		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		cmd := parts[0]
		args := parts[1:]

		switch cmd {
		case "ping":
			fmt.Fprintln(conn, "OK")
		case "identities":
			if len(args) < 1 {
				fmt.Fprintln(conn, "ERR missing identities")
				continue
			}
			ids, err := age.ParseIdentities(strings.NewReader(strings.Join(args, "\n")))
			if err != nil {
				fmt.Fprintln(conn, "ERR failed to parse identities: "+err.Error())
				continue
			}
			a.identities = ids
			fmt.Fprintln(conn, "OK")
		case "decrypt":
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
				fmt.Fprintln(conn, "ERR failed to decrypt: "+err.Error())
				continue
			}
			fmt.Fprintln(conn, "OK "+base64.StdEncoding.EncodeToString(plaintext))
		case "remove":
			if len(args) != 1 {
				fmt.Fprintln(conn, "ERR missing key")

				continue
			}
			a.cache.Remove(args[0])
			fmt.Fprintln(conn, "OK")
		case "lock":
			a.cache.Purge()
			fmt.Fprintln(conn, "OK")
		case "quit":
			fmt.Fprintln(conn, "OK")
			go a.Shutdown(ctx)

			return
		default:
			fmt.Fprintln(conn, "ERR unknown command")
		}
	}
}

func (a *Agent) decrypt(ciphertext []byte) ([]byte, error) {
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

func (a *Agent) getPassphrase(reason string, repeat bool) (string, error) {
	opts := []pinentry.ClientOption{
		pinentry.WithDesc(strings.TrimSuffix(reason, ":") + "."),
		pinentry.WithGPGTTY(),
		pinentry.WithPrompt("Passphrase:"),
		pinentry.WithTitle("gopass"),
	}
	if binary := os.Getenv("GOPASS_PINENTRY"); binary != "" {
		opts = append(opts, pinentry.WithBinaryName(binary))
	} else {
		opts = append(opts, pinentry.WithBinaryNameFromGnuPGAgentConf())
	}
	if repeat {
		opts = append(opts, pinentry.WithRepeat("Confirm"))
	} else {
		opts = append(opts,
			pinentry.WithOption(pinentry.OptionAllowExternalPasswordCache),
			pinentry.WithKeyInfo("gopass/age-identities"),
		)
	}

	p, err := pinentry.NewClient(opts...)
	if err != nil {
		debug.Log("Pinentry not found: %q", err)
		// use CLI fallback
		pf := cli.New()
		if repeat {
			_ = pf.Set("REPEAT")
		}

		return pf.GetPIN()
	}
	defer func() {
		_ = p.Close()
	}()

	result, err := p.GetPIN()
	if err != nil {
		return "", fmt.Errorf("pinentry error: %w", err)
	}

	return result.PIN, nil
}
