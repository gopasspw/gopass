package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/pkg/errors"
)

// Client is a agent client
type Client struct {
	http http.Client
}

// New creates a new client
func New(dir string) *Client {
	socket := filepath.Join(dir, ".gopass-agent.sock")
	return &Client{
		http: http.Client{
			Transport: &http.Transport{
				DialContext: func(context.Context, string, string) (net.Conn, error) {
					return net.Dial("unix", socket)
				},
			},
			Timeout: 30 * time.Second,
		},
	}
}

// Ping checks connectivity to the agent
func (c *Client) Ping() error {
	resp, err := c.http.Get("http://unix/ping")
	if err != nil {
		return err
	}
	_ = resp.Body.Close()
	return nil
}

func (c *Client) waitForAgent() error {
	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = 60 * time.Second
	return backoff.Retry(c.Ping, bo)
}

// Remove un-caches a single key
func (c *Client) Remove(key string) error {
	u, err := url.Parse("http://unix/cache/remove")
	if err != nil {
		return errors.Wrapf(err, "failed to build request url")
	}

	values := u.Query()
	values.Set("key", key)
	u.RawQuery = values.Encode()

	resp, err := c.http.Get(u.String())
	if err != nil {
		return errors.Wrapf(err, "failed to talk to agent")
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed: %d", resp.StatusCode)
	}

	return nil
}

// Passphrase asks for a passphrase from the agent
func (c *Client) Passphrase(key, reason string) (string, error) {
	if err := c.Ping(); err != nil {
		if err := c.startAgent(); err != nil {
			return "", errors.Wrapf(err, "failed to start agent")
		}
		if err := c.waitForAgent(); err != nil {
			return "", errors.Wrapf(err, "failed to start agent (expired)")
		}
	}

	u, err := url.Parse("http://unix/passphrase")
	if err != nil {
		return "", errors.Wrapf(err, "failed to build request url")
	}
	values := u.Query()
	values.Set("key", key)
	values.Set("reason", reason)
	u.RawQuery = values.Encode()

	resp, err := c.http.Get(u.String())
	if err != nil {
		return "", errors.Wrapf(err, "failed to talk to agent")
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request failed: %d", resp.StatusCode)
	}

	buf := &bytes.Buffer{}
	if _, err := io.Copy(buf, resp.Body); err != nil {
		return "", errors.Wrapf(err, "failed to talk to agent")
	}
	return buf.String(), nil
}
