package agent

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"net"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gopasspw/gopass/pkg/appdir"
	"github.com/gopasspw/gopass/pkg/debug"
)

// Client is a client for the age agent.
type Client struct {
	socketPath string
}

// NewClient creates a new client.
func NewClient() *Client {
	return &Client{
		socketPath: filepath.Join(appdir.UserRuntime(), socketName),
	}
}

func (c *Client) connect() (net.Conn, error) {
	if err := c.checkSocketSecurity(); err != nil {
		return nil, err
	}

	debug.Log("connecting to agent at %s", c.socketPath)
	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to agent: %w", err)
	}

	debug.Log("connected to agent at %s", c.socketPath)
	return conn, nil
}

func (c *Client) send(cmd string) (string, error) {
	conn, err := c.connect()
	if err != nil {
		return "", err
	}
	defer func() {
		_ = conn.Close()
	}()

	if _, err := fmt.Fprintln(conn, cmd); err != nil {
		return "", fmt.Errorf("failed to send command to agent: %w", err)
	}

	resp, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read response from agent: %w", err)
	}

	resp = strings.TrimSpace(resp)
	if strings.HasPrefix(resp, "ERR") {
		return "", fmt.Errorf("agent error: %s", strings.TrimPrefix(resp, "ERR "))
	}

	return strings.TrimPrefix(resp, "OK "), nil
}

// Ping pings the agent.
func (c *Client) Ping() error {
	_, err := c.send("ping")

	return err
}

// Status returns the agent's status.
func (c *Client) Status() (string, error) {
	return c.send("status")
}

// SendIdentities sends the identities to the agent.
func (c *Client) SendIdentities(ids string) error {
	_, err := c.send("identities " + ids)

	return err
}

// Decrypt decrypts the given ciphertext.
func (c *Client) Decrypt(ciphertext []byte) ([]byte, error) {
	resp, err := c.send("decrypt " + base64.StdEncoding.EncodeToString(ciphertext))
	if err != nil {
		return nil, err
	}

	return base64.StdEncoding.DecodeString(resp)
}

// Remove removes a passphrase from the agent.
func (c *Client) Remove(key string) error {
	_, err := c.send("remove " + key)

	return err
}

// Lock locks the agent.
func (c *Client) Lock() error {
	_, err := c.send("lock")

	return err
}

// Unlock unlocks the agent.
func (c *Client) Unlock() error {
	_, err := c.send("unlock")

	return err
}

// SetTimeout sets the agent's timeout.
func (c *Client) SetTimeout(timeout int) error {
	_, err := c.send("set-timeout " + strconv.Itoa(timeout))

	return err
}

// Quit quits the agent.
func (c *Client) Quit() error {
	_, err := c.send("quit")

	return err
}
