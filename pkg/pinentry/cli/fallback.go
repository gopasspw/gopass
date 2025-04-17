// Package cli provides a pinentry client that uses the terminal
// for input and output. It is a drop-in replacement for the
// pinentry program. It is used to ask for a passphrase or PIN
// in the terminal.
package cli

import (
	"context"
	"fmt"

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

// GetPINContext prompts for the pin in the termnial and returns the output.
// The context is only used for tests.
func (c *Client) GetPINContext(ctx context.Context) (string, error) {
	pw, err := termio.AskForPassword(ctx, "your PIN", c.repeat)
	if err != nil {
		return "", fmt.Errorf("failed to ask for PIN: %w", err)
	}

	return pw, nil
}

// GetPIN prompts for the pin in the termnial and returns the output.
func (c *Client) GetPIN() (string, error) {
	return c.GetPINContext(context.TODO())
}
