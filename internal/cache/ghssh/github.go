package ghssh

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gopasspw/gopass/pkg/debug"
)

// ListKeys returns the public keys for a github user. It will
// cache results up the a configurable amount of time (default: 6h).
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
	c.disk.Set(user, keys)
	return keys, nil
}

// fetchKeys returns the public keys for a github user.
func (c *Cache) fetchKeys(ctx context.Context, user string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	url := fmt.Sprintf("https://github.com/%s.keys", user)
	debug.Log("fetching public keys for %s from github: %s", user, url)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	out := make([]string, 0, 5)
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		out = append(out, strings.TrimSpace(scanner.Text()))
	}
	return out, nil
}
