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
	"os"
	"runtime"
	"testing"

	"github.com/gopasspw/gopass/internal/gptest"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/updater"
	"github.com/gopasspw/gopass/pkg/ctxutil"

	_ "github.com/gopasspw/gopass/internal/backend/crypto"
	_ "github.com/gopasspw/gopass/internal/backend/rcs"
	_ "github.com/gopasspw/gopass/internal/backend/storage"

	"github.com/dominikschulz/github-releases/ghrel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

const testUpdateJSON = `[
  {
    "id": 8979832,
    "name": "0.0.1 / 2017-12-02",
    "tag_name": "v0.0.1",
    "draft": false,
    "prerelease": false,
    "published_at": "2017-12-02T14:38:21Z",
    "assets": [
      {
	"browser_download_url": "%s/gopass.tar.gz",
	"id": 5676622,
	"name": "gopass-0.0.1-%s-%s.tar.gz"
      }
    ]
  },
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
	updater.UpdateMoveAfterQuit = false

	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = out.WithHidden(ctx, true)
	act, err := newMock(ctx, u)
	require.NoError(t, err)

	app := cli.NewApp()
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	c := cli.NewContext(app, fs, nil)
	c.Context = ctx

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
			Typeflag: tar.TypeReg,
			Name:     "gopass",
			Mode:     0600,
			Size:     int64(len(body)),
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
		json := fmt.Sprintf(testUpdateJSON, ghdl.URL, runtime.GOOS, runtime.GOARCH, ghdl.URL, runtime.GOOS, runtime.GOARCH)
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

	assert.NoError(t, act.Update(c))
	t.Logf("Output: %s", buf.String())
	buf.Reset()
}
