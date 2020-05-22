package updater

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/termio"

	"github.com/blang/semver"
	"github.com/cenkalti/backoff"
	"github.com/dominikschulz/github-releases/ghrel"
	"github.com/muesli/goprogressbar"
	"github.com/pkg/errors"
)

var (
	// UpdateMoveAfterQuit is exported for testing
	UpdateMoveAfterQuit = true
)

const (
	gitHubOrg  = "gopasspw"
	gitHubRepo = "gopass"
)

// Update will start hte interactive update assistant
func Update(ctx context.Context, pre bool, version semver.Version) error {
	if err := IsUpdateable(ctx); err != nil {
		out.Error(ctx, "Your gopass binary is externally managed. Can not update.")
		out.Debug(ctx, "Error: %s", err)
		return nil
	}

	ok, err := termio.AskForBool(ctx, "Do you want to check for available updates?", true)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	r, err := LatestRelease(ctx, pre || len(version.Pre) > 0)
	if err != nil {
		return err
	}

	out.Debug(ctx, "Current: %s - Latest: %s", version.String(), r.Version().String())
	if version.GTE(r.Version()) {
		out.Green(ctx, "gopass is up to date (%s)", version.String())
		return nil
	}

	out.Debug(ctx, "Assets: %+v", r.Assets)
	for _, asset := range r.Assets {
		name := strings.TrimSuffix(strings.TrimPrefix(asset.Name, "gopass-"), ".tar.gz")
		p := strings.Split(name, "-")
		if len(p) < 3 {
			continue
		}
		if p[len(p)-2] != runtime.GOOS {
			continue
		}
		if p[len(p)-1] != runtime.GOARCH {
			continue
		}
		if asset.URL == "" {
			continue
		}
		if err := updateTo(ctx, r.Version().String(), asset.URL); err != nil {
			return errors.Wrapf(err, "Failed to update gopass: %s", err)
		}
		return nil
	}
	return errors.New("no supported binary found")
}

// LatestRelease fetches and return the latest release of gopass from GitHub
func LatestRelease(ctx context.Context, pre bool) (ghrel.Release, error) {
	if pre {
		return ghrel.FetchLatestRelease(gitHubOrg, gitHubRepo)
	}
	return ghrel.FetchLatestStableRelease(gitHubOrg, gitHubRepo)
}

func updateCheckHost(ctx context.Context, u *url.URL) error {
	host, _, err := net.SplitHostPort(u.Host)
	if err != nil {
		out.Debug(ctx, "failed to split host port: %s", err)
		if e, ok := err.(*net.AddrError); ok && e.Err != "missing port in address" {
			return errors.Wrapf(err, "failed to split host port")
		}
		host = u.Host
	}
	if u.Scheme != "https" && host != "localhost" && host != "127.0.0.1" {
		return errors.Errorf("refusing non-https URL '%s'", u.String())
	}
	return nil
}

func updateTo(ctx context.Context, version, url string) error {
	out.Debug(ctx, "URL: %s", url)
	out.Green(ctx, "Update available!")
	ok, err := termio.AskForBool(ctx, fmt.Sprintf("Do you want to update gopass to %s?", version), true)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	return updateGopass(ctx, version, url)
}

func extract(ctx context.Context, archive, dest string) error {
	out.Debug(ctx, "Reading from %s", archive)
	fh, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer func() {
		_ = fh.Close()
	}()

	dfh, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0755)
	if err != nil {
		return errors.Wrapf(err, "Failed to open file: %s", dest)
	}
	defer func() {
		_ = dfh.Close()
	}()

	gzr, err := gzip.NewReader(fh)
	if err != nil {
		return err
	}
	tarReader := tar.NewReader(gzr)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return errors.Wrapf(err, "Failed to read from tar file")
		}
		name := filepath.Base(header.Name)
		if header.Typeflag == tar.TypeReg && name == "gopass" {
			_, err := io.Copy(dfh, tarReader)
			return errors.Wrapf(err, "Failed to read gopass from tar file")
		}
	}
	return errors.Errorf("file not found in archive")
}

func tryDownload(ctx context.Context, dest, url string) error {
	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = 5 * time.Minute

	return backoff.Retry(func() error {
		select {
		case <-ctx.Done():
			return backoff.Permanent(errors.New("user aborted"))
		default:
		}
		return download(ctx, dest, url)
	}, bo)
}

func download(ctx context.Context, dest, url string) error {
	fh, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0755)
	if err != nil {
		return err
	}
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	var body io.ReadCloser
	if resp.ContentLength > 0 {
		body = &passThru{
			ReadCloser: resp.Body,
			Bar: &goprogressbar.ProgressBar{
				Text:    path.Base(url),
				Total:   resp.ContentLength,
				Current: 0,
				Width:   80,
				PrependTextFunc: func(p *goprogressbar.ProgressBar) string {
					return fmt.Sprintf("%d / %d byte", p.Current, p.Total)
				},
			},
		}
		if out.IsHidden(ctx) {
			old := goprogressbar.Stdout
			goprogressbar.Stdout = ioutil.Discard
			defer func() {
				goprogressbar.Stdout = old
			}()
		}

	} else {
		body = resp.Body
	}
	count, err := io.Copy(fh, body)
	if err != nil {
		return err
	}
	fmt.Fprintln(goprogressbar.Stdout, "")
	out.Debug(ctx, "Transferred %d bytes from %s to %s", count, url, dest)
	return nil
}

type passThru struct {
	io.ReadCloser
	Bar *goprogressbar.ProgressBar
}

func (pt *passThru) Read(p []byte) (int, error) {
	n, err := pt.ReadCloser.Read(p)
	if pt.Bar != nil {
		pt.Bar.Current += int64(n)
		pt.Bar.LazyPrint()
	}
	return n, err
}
