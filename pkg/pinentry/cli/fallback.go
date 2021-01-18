package cli

import (
	"context"

	"github.com/gopasspw/gopass/pkg/termio"
)

// Client is pinentry CLI drop-in
type Client struct{}

// New creates a new client
func New() (*Client, error) {
	return &Client{}, nil
}

// Close is a no-op
func (c *Client) Close() {}

// Confirm is a no-op
func (c *Client) Confirm() bool {
	return true
}

// Set is a no-op
func (c *Client) Set(string, string) error {
	return nil
}

// Option is a no-op
func (c *Client) Option(string) error {
	return nil
}

// GetPin prompts for the pin in the termnial and returns the output
func (c *Client) GetPin() ([]byte, error) {
	pw, err := termio.AskForPassword(context.TODO(), "Please enter your PIN")
	if err != nil {
		return nil, err
	}
	return []byte(pw), nil
}
