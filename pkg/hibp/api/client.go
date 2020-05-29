// Package api implements an HIBP API client.
package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/gopasspw/gopass/internal/debug"
	"github.com/pkg/errors"
)

// URL is the HIBPv2 API URL
var URL = "https://api.pwnedpasswords.com"

// Lookup performs a lookup against the HIBP v2 API
func Lookup(shaSum string) (uint64, error) {
	if len(shaSum) != 40 {
		return 0, errors.Errorf("invalid shasum")
	}

	shaSum = strings.ToUpper(shaSum)
	prefix := shaSum[:5]
	suffix := shaSum[5:]

	var count uint64
	url := fmt.Sprintf("%s/range/%s", URL, prefix)

	op := func() error {
		debug.Log("[%s] HTTP Request: %s", shaSum, url)
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		defer func() {
			_ = resp.Body.Close()
		}()

		if resp.StatusCode == http.StatusNotFound {
			return nil
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("HTTP request failed: %s %s", resp.Status, body)
		}

		for _, line := range strings.Split(string(body), "\n") {
			line = strings.TrimSpace(line)
			if len(line) < 37 {
				continue
			}
			if line[:35] != suffix {
				continue
			}
			if iv, err := strconv.ParseUint(line[36:], 10, 64); err == nil {
				count = iv
				return nil
			}
		}
		return nil
	}

	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = 10 * time.Second

	err := backoff.Retry(op, bo)
	return count, err
}
