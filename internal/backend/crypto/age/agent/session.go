package agent

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"strconv"
	"time"
)

// handleSession handles the optional source-bound protocol. Legacy commands
// remain supported, but cannot establish a source-bound session.
func (a *Agent) handleSession(conn io.Writer, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(conn, "ERR missing session command")

		return
	}
	var err error
	switch args[0] {
	case "begin":
		if len(args) != 1 {
			err = fmt.Errorf("invalid session begin arguments")

			break
		}
		token := a.beginSession()
		if token == "" {
			err = fmt.Errorf("agent is stopped")

			break
		}
		fmt.Fprintln(conn, "OK "+token)

		return
	case "load-generation":
		err = a.loadSessionGeneration(args[1:])
	case "load":
		err = a.loadSession(args[1:])
	case "decrypt":
		var plaintext []byte
		plaintext, err = a.decryptSession(args[1:])
		if err == nil {
			fmt.Fprintln(conn, "OK "+base64.StdEncoding.EncodeToString(plaintext))

			return
		}
	default:
		err = fmt.Errorf("unknown session command")
	}
	if err != nil {
		fmt.Fprintln(conn, "ERR "+err.Error())

		return
	}
	fmt.Fprintln(conn, "OK")
}

func (a *Agent) loadSession(args []string) error {
	a.mux.Lock()
	defer a.mux.Unlock()

	return a.loadSessionLocked(args)
}

func (a *Agent) beginSession() string {
	a.mux.Lock()
	defer a.mux.Unlock()
	if a.stopped {
		return ""
	}
	a.clearLocked()
	a.unlockToken = rand.Text()

	return a.unlockToken
}

func (a *Agent) loadSessionGeneration(args []string) error {
	a.mux.Lock()
	defer a.mux.Unlock()
	if a.stopped {
		return fmt.Errorf("agent is stopped")
	}
	if len(args) != 4 {
		return fmt.Errorf("invalid session load arguments")
	}
	if args[2] == "" || args[2] != a.unlockToken {
		// Do not disturb a newer successful load or authentication attempt.
		return fmt.Errorf("session unlock invalidated by lock or reload")
	}

	return a.loadSessionLocked([]string{args[0], args[1], args[3]})
}

// loadSessionLocked requires a.mux to be held.
func (a *Agent) loadSessionLocked(args []string) error {
	if a.stopped {
		return fmt.Errorf("agent is stopped")
	}
	// Explicit reload replaces credentials, never leaving a previous session
	// usable if decoding or parsing the new identity material fails.
	a.clearLocked()
	if len(args) != 3 || args[0] == "" {
		return fmt.Errorf("invalid session load arguments")
	}
	timeout, err := strconv.Atoi(args[1])
	if err != nil || timeout < 0 {
		return fmt.Errorf("invalid session timeout")
	}
	data, err := base64.StdEncoding.DecodeString(args[2])
	if err != nil {
		return fmt.Errorf("failed to decode session identities")
	}
	defer clear(data)
	ids, err := parseIdentities(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to parse session identities: %w", err)
	}
	if len(ids) == 0 {
		return fmt.Errorf("no identities specified")
	}
	a.identities = ids
	a.source = args[0]
	a.locked = false
	a.timeout = time.Duration(timeout) * time.Second
	a.startTimerLocked()

	return nil
}

func (a *Agent) decryptSession(args []string) ([]byte, error) {
	if len(args) != 2 || args[0] == "" {
		return nil, fmt.Errorf("invalid session decrypt arguments")
	}
	ciphertext, err := base64.StdEncoding.DecodeString(args[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode ciphertext")
	}
	a.mux.Lock()
	defer a.mux.Unlock()
	if a.source != args[0] {
		a.clearLocked()
	}

	return a.decryptLocked(ciphertext)
}

// clearLocked invalidates pending authentication and releases cached identity
// references for explicit locks, timeouts and session invalidation. It does not
// guarantee immediate zeroization of all private-key copies in Go memory.
// Requires a.mux to be held.
func (a *Agent) clearLocked() {
	a.generation++
	a.unlockToken = ""
	a.identities = nil
	a.source = ""
	a.locked = true
	a.lockAt = time.Time{}
	if a.timer != nil {
		a.timer.Stop()
		a.timer = nil
	}
}

// startTimerLocked requires a.mux to be held. A superseded timeout must not
// clear credentials loaded by a newer session.
func (a *Agent) startTimerLocked() {
	if a.timer != nil {
		a.timer.Stop()
		a.timer = nil
	}
	a.generation++
	a.unlockToken = ""
	generation := a.generation
	if a.timeout <= 0 {
		a.lockAt = time.Time{}

		return
	}
	a.lockAt = time.Now().Add(a.timeout)
	a.timer = time.AfterFunc(a.timeout, func() {
		a.mux.Lock()
		defer a.mux.Unlock()
		if a.generation == generation {
			a.clearLocked()
		}
	})
}
