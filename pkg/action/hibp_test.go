package action

import (
	"bytes"
	"compress/gzip"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gopasspw/gopass/pkg/ctxutil"
	hibpapi "github.com/gopasspw/gopass/pkg/hibp/api"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/tests/gptest"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

const testHibpSample = `000000005AD76BD555C1D6D771DE417A4B87E4B4
00000000A8DAE4228F821FB418F59826079BF368:42
00000000DD7F2A1C68A35673713783CA390C9E93:42
00000001E225B908BAC31C56DB04D892E47536E0:42
00000008CD1806EB7B9B46A8F87690B2AC16F617:42
0000000A0E3B9F25FF41DE4B5AC238C2D545C7A8:42
0000000A1D4B746FAA3FD526FF6D5BC8052FDB38:42
0000000CAEF405439D57847A8657218C618160B2:42
0000000FC1C08E6454BED24F463EA2129E254D43:42
00000010F4B38525354491E099EB1796278544B1`

func TestHIBPDump(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = out.WithHidden(ctx, true)
	act, err := newMock(ctx, u)
	assert.NoError(t, err)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	app := cli.NewApp()
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	c := cli.NewContext(app, fs, nil)

	// no hibp dump, no env var
	assert.Error(t, act.HIBP(ctx, c))
	buf.Reset()

	// setup file and env
	fn := filepath.Join(u.Dir, "dump.txt")
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	bf := cli.StringSliceFlag{
		Name:  "dumps",
		Usage: "dumps",
	}
	assert.NoError(t, bf.ApplyWithError(fs))
	assert.NoError(t, fs.Parse([]string{"--dumps=" + fn}))
	c = cli.NewContext(app, fs, nil)

	assert.NoError(t, ioutil.WriteFile(fn, []byte(testHibpSample), 0644))
	assert.NoError(t, act.HIBP(ctx, c))
	buf.Reset()

	// gzip
	fn = filepath.Join(u.Dir, "dump.txt.gz")
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	bf = cli.StringSliceFlag{
		Name:  "dumps",
		Usage: "dumps",
	}
	assert.NoError(t, bf.ApplyWithError(fs))
	assert.NoError(t, fs.Parse([]string{"--dumps=" + fn}))
	c = cli.NewContext(app, fs, nil)
	assert.NoError(t, testWriteGZ(fn, []byte(testHibpSample)))
	assert.NoError(t, act.HIBP(ctx, c))
	buf.Reset()
}

func testWriteGZ(fn string, buf []byte) error {
	fh, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() {
		_ = fh.Close()
	}()

	gzw := gzip.NewWriter(fh)
	defer func() {
		_ = gzw.Close()
	}()

	_, err = gzw.Write(buf)
	return err
}

func TestHIBPAPI(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = out.WithHidden(ctx, true)
	act, err := newMock(ctx, u)
	assert.NoError(t, err)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	app := cli.NewApp()
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	bf := cli.BoolFlag{
		Name:  "api",
		Usage: "api",
	}
	assert.NoError(t, bf.ApplyWithError(fs))
	assert.NoError(t, fs.Parse([]string{"--api=true"}))
	c := cli.NewContext(app, fs, nil)

	reqCnt := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqCnt++
		if reqCnt < 2 {
			http.Error(w, "fake error", http.StatusInternalServerError)
			return
		}
		if strings.TrimPrefix(r.URL.String(), "/range/") == "8843D" {
			fmt.Fprintf(w, "8843D:1\n")                                     // invalid
			fmt.Fprintf(w, "7F92416211DE9EBB963FF4CE2812593287:3234879\n")  // invalid
			fmt.Fprintf(w, "7F92416211DE9EBB963FF4CE28125932878:\n")        // invalid
			fmt.Fprintf(w, "7F92416211DE9EBB963FF4CE28125932878\n")         // invalid
			fmt.Fprintf(w, "7F92416211DE9EBB963FF4CE28125932878:3234879\n") // valid
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer ts.Close()
	hibpapi.URL = ts.URL

	// test with one entry
	assert.NoError(t, act.HIBP(ctx, c))
	buf.Reset()

	// add another one
	assert.NoError(t, act.insertStdin(ctx, "baz", []byte("foobar"), false))
	assert.Error(t, act.HIBP(ctx, c))
	buf.Reset()
}
