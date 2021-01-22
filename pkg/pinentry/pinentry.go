// Package pinentry implements a cross platform pinentry client. It can be used
// to obtain credentials from the user through a simple UI application.
package pinentry

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/pkg/errors"
)

var (
	// Unescape enables unescaping of percent encoded values,
	// disabled by default for backward compatibility. See #1621
	Unescape bool
)

func init() {
	if sv := os.Getenv("GOPASS_PINENTRY_UNESCAPE"); sv == "on" || sv == "true" {
		Unescape = true
	}
}

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

// Option sets an option, e.g. "default-cancel=Abbruch" or "allow-external-password-cache"
func (c *Client) Option(value string) error {
	val := "OPTION " + value + "\n"
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
	buf, _, err := c.out.ReadLine()
	if err != nil {
		return nil, err
	}
	if bytes.HasPrefix(buf, []byte("OK")) {
		return nil, nil
	}
	// handle status messages
	for bytes.HasPrefix(buf, []byte("S ")) {
		debug.Log("message: %q", string(buf))
		buf, _, err = c.out.ReadLine()
		if err != nil {
			return nil, err
		}
	}
	// now there should be some data
	if !bytes.HasPrefix(buf, []byte("D ")) {
		return nil, fmt.Errorf("unexpected response: %s", buf)
	}

	pin := make([]byte, len(buf))
	if n := copy(pin, buf); n != len(buf) {
		return nil, fmt.Errorf("failed to copy pin: expected %d bytes only copied %d", len(buf), n)
	}

	ok, _, err := c.out.ReadLine()
	if err != nil {
		return nil, err
	}
	if !bytes.HasPrefix(ok, []byte("OK")) {
		return nil, fmt.Errorf("unexpected response: %s", ok)
	}
	pin = pin[2:]

	// Handle unescaping, if enabled
	if bytes.Contains(pin, []byte("%")) && Unescape {
		sv, err := url.QueryUnescape(string(pin))
		if err != nil {
			return nil, fmt.Errorf("failed to unescape pin: %q", err)
		}
		pin = []byte(sv)
	}
	return pin, nil
}
