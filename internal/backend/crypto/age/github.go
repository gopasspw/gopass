package age

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func (a *Age) getPublicKeysGithub(ctx context.Context, user string) ([]string, error) {
	// TODO: recheck SoT if cache is too old
	pk, err := a.ghCache.Get(user)
	if err != nil {
		return nil, err
	}
	if len(pk) > 0 {
		return pk, nil
	}

	keys, err := githubListKeys(ctx, user)
	if err != nil {
		return nil, err
	}
	if len(keys) < 1 {
		return nil, fmt.Errorf("not found")
	}
	a.ghCache.Set(user, keys)
	return keys, nil
}

func githubListKeys(ctx context.Context, user string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://github.com/%s.keys", user), nil)
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
