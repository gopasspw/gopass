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

// Update will start th interactive update assistant
func Update(ctx context.Context, pre bool, version semver.Version, migrationCheck func(context.Context) bool) error {
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

	rs, err := FetchReleases(ctx, pre || len(version.Pre) > 0)
	if err != nil {
		return err
	}
	if len(rs) < 1 {
		return fmt.Errorf("no releases available")
	}

	out.Debug(ctx, "Current: %s - Latest: %s", version.String(), rs[0].Version().String())
	// binary is newer or equal to the latest release -> nothing to do
	if version.GTE(rs[0].Version()) {
		out.Green(ctx, "gopass is up to date (%s)", version.String())
		return nil
	}

	// binary has the same major version as the latest release -> simple update
	//if version.Major == rs[0].Version().Major || version.Major == 0 {
	if version.Major == rs[0].Version().Major {
		return simpleUpdate(ctx, rs[0])
	}

	// binary has a previous major version -> need to update to the latest
	// release of the previous minor version first, run update check and then
	// update to the next major version.
	latestMinorReleases := filterMajor(rs, version.Major)
	if len(latestMinorReleases) < 1 {
		return fmt.Errorf("no suitable releases")
	}

	// we're already at the latest release of the previous stable release
	// cycle. We attempt to migrate the any outdated data of config and
	// if the succeeds we can go straight to the latest release.
	if version.EQ(latestMinorReleases[0].Version()) && migrationCheck(ctx) {
		return simpleUpdate(ctx, rs[0])
	}

	// before we can move to the next major release we first need to update to
	// the latest release of the current major release and pass the migration
	// check.
	return simpleUpdate(ctx, latestMinorReleases[0])
}

func filterMajor(rs []ghrel.Release, major uint64) []ghrel.Release {
	out := make([]ghrel.Release, 0, len(rs)-1)
	for _, r := range rs {
		v := r.Version()
		if v.Major != major {
			continue
		}
		out = append(out, r)
	}
	return out
}

func simpleUpdate(ctx context.Context, r ghrel.Release) error {
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

// FetchReleases fetches and returns all releases of gopass from GitHub
func FetchReleases(ctx context.Context, pre bool) ([]ghrel.Release, error) {
	if pre {
		return ghrel.FetchAllReleases(gitHubOrg, gitHubRepo)
	}
	return ghrel.FetchAllStableReleases(gitHubOrg, gitHubRepo)
}

// LatestRelease fetches and returns the latest gopass release from GitHub
func LatestRelease(ctx context.Context, pre bool) (ghrel.Release, error) {
	rs, err := FetchReleases(ctx, pre)
	if err != nil {
		return ghrel.Release{}, err
	}
	if len(rs) < 1 {
		return ghrel.Release{}, fmt.Errorf("no releases")
	}
	return rs[0], nil
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
