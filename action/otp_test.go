package action

import (
	"bytes"
	"context"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/gokyle/twofactor"
	"github.com/justwatchcom/gopass/store/secret"
	"github.com/justwatchcom/gopass/tests/gptest"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestOTP(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	act, err := newMock(ctx, u)
	assert.NoError(t, err)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	app := cli.NewApp()

	// display non-otp secret
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"foo"}))
	c := cli.NewContext(app, fs, nil)

	assert.Error(t, act.OTP(ctx, c))
	buf.Reset()

	// create and display valid OTP
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"bar"}))
	c = cli.NewContext(app, fs, nil)

	assert.NoError(t, act.Store.Set(ctx, "bar", secret.New("foo", twofactor.GenerateGoogleTOTP().URL("foo"))))

	assert.NoError(t, act.OTP(ctx, c))
	buf.Reset()

	// write QR file
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	sf := cli.StringFlag{
		Name:  "qr",
		Usage: "qr",
	}
	assert.NoError(t, sf.ApplyWithError(fs))
	fn := filepath.Join(u.Dir, "qr.png")
	assert.NoError(t, fs.Parse([]string{"--qr=" + fn, "bar"}))
	c = cli.NewContext(app, fs, nil)

	assert.NoError(t, act.OTP(ctx, c))
	assert.FileExists(t, fn)
	buf.Reset()
}
