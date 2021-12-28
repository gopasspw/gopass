package action

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"testing"

	_ "github.com/gopasspw/gopass/internal/backend/crypto"
	_ "github.com/gopasspw/gopass/internal/backend/storage"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/updater"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testUpdateJSON = `{
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
      },
      {
       "browser_download_url": "%s/SHA256SUMS",
       "id": 5676624,
       "name": "gopass-1.6.6_SHA256SUMS"
      },
      {
       "browser_download_url": "%s/SHA256SUMS.sig",
       "id": 5676625,
       "name": "gopass-1.6.6_SHA256SUMS.sig"
      }
    ]
  }`

func TestUpdate(t *testing.T) {
	updater.UpdateMoveAfterQuit = false

	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithHidden(ctx, true)
	act, err := newMock(ctx, u)
	require.NoError(t, err)

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
		json := fmt.Sprintf(testUpdateJSON, ghdl.URL, runtime.GOOS, runtime.GOARCH, ghdl.URL, ghdl.URL)
		fmt.Fprint(w, json)
	}))
	defer ghapi.Close()

	updater.BaseURL = ghapi.URL + "/%s/%s"

	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		stdout = os.Stdout
	}()

	// TODO: This should not fail, but then we need to provide valid signatures
	assert.Error(t, act.Update(gptest.CliCtx(ctx, t)))
	buf.Reset()
}
