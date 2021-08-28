package cli

import (
	"context"

	"github.com/gopasspw/gopass/pkg/termio"
)

// Client is pinentry CLI drop-in
type Client struct {
	repeat bool
}

// New creates a new client
func New() (*Client, error) {
	return &Client{repeat: false}, nil
}

// Close is a no-op
func (c *Client) Close() {}

// Confirm is a no-op
func (c *Client) Confirm() bool {
	return true
}

// Set is a no-op unless you're requesting a repeat
func (c *Client) Set(key string, _ string) error {
	if key == "REPEAT" {
		c.repeat = true
	}
	return nil
}

// Option is a no-op
func (c *Client) Option(string) error {
	return nil
}

// GetPin prompts for the pin in the termnial and returns the output
func (c *Client) GetPin() ([]byte, error) {
	pw, err := termio.AskForPassword(context.TODO(), "your PIN", c.repeat)
	if err != nil {
		return nil, err
	}
	return []byte(pw), nil
}
