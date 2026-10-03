package agent

import (
	"bufio"
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"filippo.io/age"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

func TestSourceBoundSession(t *testing.T) {
	// Use a private socket without Run's process-wide signal handler or runtime
	// directory so this test cannot contact the user's agent.
	path := filepath.Join(t.TempDir(), "agent.sock")
	listener, err := net.Listen("unix", path)
	require.NoError(t, err)
	require.NoError(t, os.Chmod(path, 0o600))
	a := &Agent{socketPath: path, listener: listener, locked: true}
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go a.handleConnection(t.Context(), conn)
		}
	}()
	t.Cleanup(func() { a.Shutdown(t.Context()) })
	c := &Client{socketPath: path}
	id, err := age.GenerateX25519Identity()
	require.NoError(t, err)
	first := sessionCiphertext(t, id, "first entry")
	second := sessionCiphertext(t, id, "second entry")

	// Only the initial keyring load is needed to decrypt multiple entries.
	require.NoError(t, c.LoadIdentities(id.String(), "keyring-v1", 0))
	for _, ciphertext := range [][]byte{first, second} {
		_, err = c.DecryptSession(ciphertext, "keyring-v1")
		require.NoError(t, err)
	}
	_, err = c.Decrypt(first)
	require.ErrorContains(t, err, "session decryption requires source")

	// A changed keyring cannot continue using cached keys, even if they would
	// still decrypt the requested entry.
	_, err = c.DecryptSession(first, "keyring-v2")
	require.ErrorContains(t, err, "agent is locked")
	a.mux.Lock()
	require.Empty(t, a.identities)
	a.mux.Unlock()
	status, err := c.Status()
	require.NoError(t, err)
	require.Equal(t, "locked", status)

	// Failed reloads erase previous credentials and do not unlock the agent.
	require.NoError(t, c.LoadIdentities(id.String(), "keyring-v2", 0))
	require.Error(t, c.LoadIdentities("invalid identity", "keyring-v2", 0))
	_, err = c.DecryptSession(first, "keyring-v2")
	require.ErrorContains(t, err, "agent is locked")
	a.mux.Lock()
	require.Empty(t, a.identities)
	a.mux.Unlock()

	require.NoError(t, c.LoadIdentities(id.String(), "keyring-v2", 0))
	require.NoError(t, c.Lock())
	_, err = c.DecryptSession(first, "keyring-v2")
	require.ErrorContains(t, err, "agent is locked")

	// Legacy credentials cannot accidentally carry over a session fingerprint.
	require.NoError(t, c.SendIdentities(id.String()))
	require.NoError(t, c.Unlock())
	_, err = c.DecryptSession(first, "keyring-v2")
	require.ErrorContains(t, err, "agent is locked")

	// A screen lock while the hardware prompt is open must invalidate its
	// eventual answer rather than let that answer reopen the session.
	generation, err := c.BeginSession()
	require.NoError(t, err)
	require.NoError(t, c.Lock())
	err = c.LoadIdentitiesWithGeneration(id.String(), "keyring-v2", 0, generation)
	require.ErrorContains(t, err, "session unlock invalidated")
	status, err = c.Status()
	require.NoError(t, err)
	require.Equal(t, "locked", status)

	// A second unlock attempt supersedes the first. A late first answer
	// cannot clear credentials loaded by the newer attempt.
	oldGeneration, err := c.BeginSession()
	require.NoError(t, err)
	newGeneration, err := c.BeginSession()
	require.NoError(t, err)
	require.NoError(t, c.LoadIdentitiesWithGeneration(id.String(), "keyring-v2", 0, newGeneration))
	err = c.LoadIdentitiesWithGeneration(id.String(), "old-keyring", 0, oldGeneration)
	require.ErrorContains(t, err, "session unlock invalidated")
	_, err = c.DecryptSession(first, "keyring-v2")
	require.NoError(t, err)

	// A generation is single-use, including after a failed payload parse.
	generation, err = c.BeginSession()
	require.NoError(t, err)
	require.Error(t, c.LoadIdentitiesWithGeneration("invalid identity", "keyring-v2", 0, generation))
	require.Error(t, c.LoadIdentitiesWithGeneration(id.String(), "keyring-v2", 0, generation))

	require.NoError(t, c.LoadIdentities(id.String(), "keyring-v2", 0))
	a.Shutdown(t.Context())
	a.mux.Lock()
	require.Empty(t, a.identities)
	a.mux.Unlock()
	require.Empty(t, a.source)
	require.True(t, a.locked)
}

func TestSessionAutoLock(t *testing.T) {
	a := &Agent{locked: true}
	id, err := age.GenerateX25519Identity()
	require.NoError(t, err)
	encoded := base64.StdEncoding.EncodeToString([]byte(id.String()))
	require.NoError(t, a.loadSession([]string{"source", "1", encoded}))
	require.Eventually(t, func() bool {
		a.mux.Lock()
		defer a.mux.Unlock()

		return a.locked && len(a.identities) == 0 && a.source == ""
	}, 3*time.Second, 10*time.Millisecond)
}

func TestSessionReloadCancelsTimeout(t *testing.T) {
	a := &Agent{locked: true}
	t.Cleanup(a.lock)
	id, err := age.GenerateX25519Identity()
	require.NoError(t, err)
	encoded := base64.StdEncoding.EncodeToString([]byte(id.String()))
	require.NoError(t, a.loadSession([]string{"old", "1", encoded}))
	require.NoError(t, a.loadSession([]string{"new", "0", encoded}))
	time.Sleep(1100 * time.Millisecond)
	a.mux.Lock()
	defer a.mux.Unlock()
	require.False(t, a.locked)
	require.Equal(t, "new", a.source)
	require.Len(t, a.identities, 1)
}

func sessionCiphertext(t *testing.T, id *age.X25519Identity, plaintext string) []byte {
	t.Helper()
	var ciphertext bytes.Buffer
	writer, err := age.Encrypt(&ciphertext, id.Recipient())
	require.NoError(t, err)
	_, err = writer.Write([]byte(plaintext))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	return ciphertext.Bytes()
}

func TestUnlockTokenDoesNotSurviveAgentRestart(t *testing.T) {
	oldAgent := &Agent{locked: true}
	token := oldAgent.beginSession()
	newAgent := &Agent{locked: true}
	newToken := newAgent.beginSession()
	require.NotEqual(t, token, newToken)
	id, err := age.GenerateX25519Identity()
	require.NoError(t, err)
	encoded := base64.StdEncoding.EncodeToString([]byte(id.String()))
	require.ErrorContains(t, newAgent.loadSessionGeneration([]string{"source", "0", token, encoded}), "session unlock invalidated")
	require.NoError(t, newAgent.loadSessionGeneration([]string{"source", "0", newToken, encoded}))
	newAgent.lock()
}

func TestAcceptedConnectionCannotReopenStoppedAgent(t *testing.T) {
	a := &Agent{locked: true}
	generation := a.beginSession()
	id, err := age.GenerateX25519Identity()
	require.NoError(t, err)
	encodedIDs := base64.StdEncoding.EncodeToString([]byte(id.String()))
	encodedCiphertext := base64.StdEncoding.EncodeToString(sessionCiphertext(t, id, "entry"))
	client, server := net.Pipe()
	t.Cleanup(func() { _ = client.Close() })
	go a.handleConnection(t.Context(), server)
	// This connection predates shutdown and remains connected after the
	// listener has closed. Neither modern nor legacy commands may reopen it.
	a.Shutdown(t.Context())
	reader := bufio.NewReader(client)
	commands := []string{
		"session begin",
		"session load-generation source 0 " + generation + " " + encodedIDs,
		"session load source 0 " + encodedIDs,
		"identities " + id.String(),
		"unlock",
		"session decrypt source " + encodedCiphertext,
		"decrypt " + encodedCiphertext,
	}
	for _, command := range commands {
		_, err := fmt.Fprintln(client, command)
		require.NoError(t, err)
		response, err := reader.ReadString('\n')
		require.NoError(t, err)
		require.Contains(t, response, "ERR ")
		require.Contains(t, response, "agent is stopped")
	}
	a.mux.Lock()
	defer a.mux.Unlock()
	require.True(t, a.stopped)
	require.True(t, a.locked)
	require.Empty(t, a.identities)
}

// The legacy SSH upload must not add untracked credentials to a source-bound
// session, supersede pending hardware authentication, or reopen a stopped agent.
func TestSSHUploadRespectsKeyringSession(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	block, err := ssh.MarshalPrivateKey(privateKey, "test session isolation")
	require.NoError(t, err)
	args := []string{base64.StdEncoding.EncodeToString(pem.EncodeToMemory(block))}
	id, err := age.GenerateX25519Identity()
	require.NoError(t, err)
	encoded := base64.StdEncoding.EncodeToString([]byte(id.String()))
	a := &Agent{locked: true}
	t.Cleanup(a.lock)
	require.NoError(t, a.loadSession([]string{"protected-source", "0", encoded}))
	require.ErrorContains(t, a.addSSHIdentity(args), "cannot add external SSH identities")
	require.Len(t, a.identities, 1)
	plaintext, err := a.decryptSession([]string{"protected-source", base64.StdEncoding.EncodeToString(sessionCiphertext(t, id, "protected"))})
	require.NoError(t, err)
	require.Equal(t, "protected", string(plaintext))

	token := a.beginSession()
	require.ErrorContains(t, a.addSSHIdentity(args), "cannot add external SSH identities")
	require.Equal(t, token, a.unlockToken)
	require.Empty(t, a.identities)
	require.NoError(t, a.loadSessionGeneration([]string{"protected-source", "0", token, encoded}))

	a.Shutdown(t.Context())
	require.ErrorContains(t, a.addSSHIdentity(args), "agent is stopped")
	require.Empty(t, a.identities)
}
