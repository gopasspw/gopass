package agent

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"net"
	"path/filepath"
	"strings"

	"github.com/gopasspw/gopass/pkg/appdir"
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
	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to agent: %w", err)
	}

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

// Quit quits the agent.
func (c *Client) Quit() error {
	_, err := c.send("quit")

	return err
}
