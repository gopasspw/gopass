package updater

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/termio"
	"golang.org/x/net/context/ctxhttp"
)

var (
	// DownloadTimeout is the overall timeout for the download, including all retries.
	DownloadTimeout = time.Minute * 5
	httpClient      = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				// enforce TLS 1.3
				MinVersion: tls.VersionTLS13,
			},
		},
	}
)

func tryDownload(ctx context.Context, url string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, DownloadTimeout)
	defer cancel()

	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = DownloadTimeout

	var buf []byte

	return buf, backoff.Retry(func() error {
		select {
		case <-ctx.Done():
			return backoff.Permanent(fmt.Errorf("user aborted"))
		default:
		}
		d, err := download(ctx, url)
		if err == nil {
			buf = d
		}
		return err
	}, bo)
}

func download(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	// we want binary data, please
	req.Header.Set("Accept", "application/octet-stream")

	t0 := time.Now()
	resp, err := ctxhttp.Do(ctx, httpClient, req)
	if err != nil {
		return nil, err
	}

	var body io.ReadCloser
	// do not show progress bar for small assets, like SHA256SUMS
	bar := termio.NewProgressBar(resp.ContentLength)
	bar.Hidden = ctxutil.IsHidden(ctx) || resp.ContentLength < 10000

	body = &passThru{
		ReadCloser: resp.Body,
		Bar:        bar,
	}

	buf := &bytes.Buffer{}
	count, err := io.Copy(buf, body)
	if err != nil {
		return nil, err
	}

	bar.Set(resp.ContentLength)
	bar.Done()

	elapsed := time.Since(t0)
	debug.Log("Transferred %d bytes from %q in %s", count, url, elapsed)
	return buf.Bytes(), nil
}

type setter interface {
	Set(int64)
}

type passThru struct {
	io.ReadCloser
	Bar setter
}

func (pt *passThru) Read(p []byte) (int, error) {
	n, err := pt.ReadCloser.Read(p)
	if pt.Bar != nil && n > 0 {
		pt.Bar.Set(int64(n))
	}
	return n, err
}
