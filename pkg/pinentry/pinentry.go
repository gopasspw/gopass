// Package pinentry implements a cross platform pinentry client. It can be used
// to obtain credentials from the user through a simple UI application.
package pinentry

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

// Client is a pinentry client
type Client struct {
	cmd *exec.Cmd
	in  io.WriteCloser
	out *bufio.Reader
}

// New creates a new pinentry client
func New() (*Client, error) {
	cmd := exec.Command(GetBinary())
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	br := bufio.NewReader(stdout)
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	// check welcome message
	banner, _, err := br.ReadLine()
	if err != nil {
		return nil, err
	}
	if !bytes.HasPrefix(banner, []byte("OK")) {
		return nil, fmt.Errorf("wrong banner: %s", banner)
	}

	cl := &Client{
		cmd: cmd,
		in:  stdin,
		out: br,
	}

	return cl, nil
}

// Close closes the client
func (c *Client) Close() {
	_ = c.in.Close()
}

// Confirm sends the confirm message
func (c *Client) Confirm() bool {
	if err := c.Set("confirm", ""); err == nil {
		return true
	}
	return false
}

// Set sets a key
func (c *Client) Set(key, value string) error {
	key = strings.ToUpper(key)
	if value != "" {
		value = " " + value
	}
	val := "SET" + key + value + "\n"
	if _, err := c.in.Write([]byte(val)); err != nil {
		return err
	}
	line, _, _ := c.out.ReadLine()
	if string(line) != "OK" {
		return errors.Errorf("error: %s", line)
	}
	return nil
}

// GetPin asks for the pin
func (c *Client) GetPin() ([]byte, error) {
	if _, err := c.in.Write([]byte("GETPIN\n")); err != nil {
		return nil, err
	}
	pin, _, err := c.out.ReadLine()
	if err != nil {
		return nil, err
	}
	if bytes.HasPrefix(pin, []byte("OK")) {
		return nil, nil
	}
	if !bytes.HasPrefix(pin, []byte("D ")) {
		return nil, fmt.Errorf("unexpected response: %s", pin)
	}

	ok, _, err := c.out.ReadLine()
	if err != nil {
		return nil, err
	}
	if !bytes.HasPrefix(ok, []byte("OK")) {
		return nil, fmt.Errorf("unexpected response: %s", ok)
	}
	return pin[2:], nil
}
