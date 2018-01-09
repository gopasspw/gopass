package action

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"runtime"
	"testing"

	"github.com/dominikschulz/github-releases/ghrel"
	"github.com/justwatchcom/gopass/tests/gptest"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

const testUpdateJSON = `[
  {
    "id": 8979833,
    "name": "1.6.6 / 2017-12-20",
    "tag_name": "v1.6.6",
    "draft": false,
    "prerelease": false,
    "published_at": "2017-12-20T14:38:21Z",
    "assets": [
      {
        "browser_download_url": "%s/gopass.tar.gz",
        "id": 5676623,
        "name": "gopass-1.6.6-%s-%s.tar.gz"
      }
    ]
  }
]`

func TestUpdate(t *testing.T) {
	updateMoveAfterQuit = false

	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = out.WithHidden(ctx, true)
	act, err := newMock(ctx, u)
	assert.NoError(t, err)

	app := cli.NewApp()
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	c := cli.NewContext(app, fs, nil)

	// github release download mock
	ghdl := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gzw := gzip.NewWriter(w)
		defer func() {
			_ = gzw.Close()
		}()
		tw := tar.NewWriter(gzw)
		defer func() {
			_ = tw.Close()
		}()
		body := "foobar"
		hdr := &tar.Header{
			Name: "gopass",
			Mode: 0600,
			Size: int64(len(body)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if _, err := tw.Write([]byte(body)); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}))
	defer ghdl.Close()
	// github api mock
	ghapi := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json := fmt.Sprintf(testUpdateJSON, ghdl.URL, runtime.GOOS, runtime.GOARCH)
		fmt.Fprint(w, json)
	}))
	defer ghapi.Close()

	ghrel.BaseURL = ghapi.URL + "/%s/%s"

	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		stdout = os.Stdout
	}()

	assert.NoError(t, act.Update(ctx, c))
	buf.Reset()
}

func TestCheckHost(t *testing.T) {
	ctx := context.Background()

	for _, tc := range []struct {
		in string
		ok bool
	}{
		{
			in: "https://github.com/justwatchcom/gopass/releases/download/v1.6.8/gopass-1.6.8-linux-amd64.tar.gz",
			ok: true,
		},
		{
			in: "http://localhost:8080/foo/bar.tar.gz",
			ok: true,
		},
	} {
		u, err := url.Parse(tc.in)
		assert.NoError(t, err)
		err = updateCheckHost(ctx, u)
		if tc.ok {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}
