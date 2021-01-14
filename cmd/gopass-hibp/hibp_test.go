package main

import (
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
	"github.com/gopasspw/gopass/pkg/gopass/apimock"
	hibpapi "github.com/gopasspw/gopass/pkg/hibp/api"
	"github.com/gopasspw/gopass/tests/gptest"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
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
	dir, err := ioutil.TempDir("", "gopass-hibp")
	if err != nil {
		t.Fatalf("failed to create temp dir: %s", err)
	}
	defer os.RemoveAll(dir)

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	act := &hibp{
		gp: apimock.New(),
	}

	app := cli.NewApp()
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	c := cli.NewContext(app, fs, nil)
	c.Context = ctx

	// setup file and env
	fn := filepath.Join(dir, "dump.txt")
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	bf := cli.StringSliceFlag{
		Name:  "dumps",
		Usage: "dumps",
	}
	assert.NoError(t, bf.Apply(fs))
	assert.NoError(t, fs.Parse([]string{"--dumps=" + fn}))
	c = cli.NewContext(app, fs, nil)
	c.Context = ctx

	assert.NoError(t, ioutil.WriteFile(fn, []byte(testHibpSample), 0644))
	assert.NoError(t, act.CheckDump(c.Context, false, []string{fn}))

	// gzip
	fn = filepath.Join(dir, "dump.txt.gz")
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	bf = cli.StringSliceFlag{
		Name:  "dumps",
		Usage: "dumps",
	}
	assert.NoError(t, bf.Apply(fs))
	assert.NoError(t, fs.Parse([]string{"--dumps=" + fn}))

	c = cli.NewContext(app, fs, nil)
	c.Context = ctx

	assert.NoError(t, testWriteGZ(fn, []byte(testHibpSample)))
	assert.NoError(t, act.CheckDump(c.Context, false, []string{fn}))
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
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	act := &hibp{
		gp: apimock.New(),
	}

	c := gptest.CliCtxWithFlags(ctx, t, map[string]string{"api": "true"})

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
	assert.NoError(t, act.CheckAPI(c.Context, false))

	// add another one
	assert.NoError(t, act.gp.Set(ctx, "baz", &apimock.Secret{Buf: []byte("foobar")}))
	assert.Error(t, act.CheckAPI(c.Context, false))
}
