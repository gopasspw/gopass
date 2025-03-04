package ghssh

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gopasspw/gopass/pkg/debug"
)

var httpClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			// enforce TLS 1.3
			MinVersion: tls.VersionTLS13,
		},
	},
}

var baseURL = "https://github.com"

// ListKeys returns the public keys for a github user. It will
// cache results up to a configurable amount of time (default: 6h).
func (c *Cache) ListKeys(ctx context.Context, user string) ([]string, error) {
	pk, err := c.disk.Get(user)
	if err != nil {
		debug.Log("failed to fetch %s from cache: %s", user, err)
	}

	if len(pk) > 0 {
		return pk, nil
	}

	keys, err := c.fetchKeys(ctx, user)
	if err != nil {
		return nil, err
	}

	if len(keys) < 1 {
		return nil, fmt.Errorf("key not found")
	}

	_ = c.disk.Set(user, keys)

	return keys, nil
}

// fetchKeys returns the public keys for a github user.
func (c *Cache) fetchKeys(ctx context.Context, user string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s/%s.keys", baseURL, user)
	debug.Log("fetching public keys for %s from github: %s", user, url)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch keys from %s: %s", url, resp.Status)
	}

	out := make([]string, 0, 5)
	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		out = append(out, strings.TrimSpace(scanner.Text()))
	}

	return out, nil
}
