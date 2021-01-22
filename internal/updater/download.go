package updater

import (
	"bytes"
	"context"
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

func tryDownload(ctx context.Context, url string) ([]byte, error) {
	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = 5 * time.Minute

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

	resp, err := ctxhttp.Do(ctx, http.DefaultClient, req)
	if err != nil {
		return nil, err
	}

	var body io.ReadCloser
	bar := termio.NewProgressBar(resp.ContentLength, ctxutil.IsHidden(ctx))
	// do not show progress bar for small assets, like SHA256SUMS
	if resp.ContentLength > 10000 {
		body = &passThru{
			ReadCloser: resp.Body,
			Bar:        bar,
		}

	} else {
		body = resp.Body
	}

	buf := &bytes.Buffer{}

	count, err := io.Copy(buf, body)
	if err != nil {
		return nil, err
	}
	bar.Done()
	debug.Log("Transferred %d bytes from %s", count, url)
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
	if pt.Bar != nil {
		pt.Bar.Set(int64(n))
	}
	return n, err
}
