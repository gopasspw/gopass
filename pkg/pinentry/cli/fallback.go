package cli

import (
	"context"

	"github.com/gopasspw/gopass/pkg/termio"
)

// Client is pinentry CLI drop-in.
type Client struct {
	repeat bool
}

// New creates a new client.
func New() *Client {
	return &Client{repeat: false}
}

// Set is a no-op unless you're requesting a repeat.
func (c *Client) Set(key string) error {
	if key == "REPEAT" {
		c.repeat = true
	}
	return nil
}

// Option is a no-op.
func (c *Client) Option(string) error {
	return nil
}

// GetPIN prompts for the pin in the termnial and returns the output.
func (c *Client) GetPIN() (string, error) {
	pw, err := termio.AskForPassword(context.TODO(), "your PIN", c.repeat)
	if err != nil {
		return "", err
	}
	return pw, nil
}
