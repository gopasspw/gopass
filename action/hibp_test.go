package action

import (
	"bytes"
	"context"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

const testHibpSample = `000000005AD76BD555C1D6D771DE417A4B87E4B4
00000000A8DAE4228F821FB418F59826079BF368
00000000DD7F2A1C68A35673713783CA390C9E93
00000001E225B908BAC31C56DB04D892E47536E0
00000008CD1806EB7B9B46A8F87690B2AC16F617
0000000A0E3B9F25FF41DE4B5AC238C2D545C7A8
0000000A1D4B746FAA3FD526FF6D5BC8052FDB38
0000000CAEF405439D57847A8657218C618160B2
0000000FC1C08E6454BED24F463EA2129E254D43
00000010F4B38525354491E099EB1796278544B1`

func TestHIBP(t *testing.T) {
	td, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = out.WithHidden(ctx, true)
	act, err := newMock(ctx, td)
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
	fn := filepath.Join(td, "dump.txt")
	assert.NoError(t, ioutil.WriteFile(fn, []byte(testHibpSample), 0644))
	assert.NoError(t, os.Setenv("HIBP_DUMPS", fn))
	assert.NoError(t, act.HIBP(ctx, c))
	buf.Reset()
}
