package action

import (
	"bytes"
	"context"
	"flag"
	"os"
	"testing"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/tests/gptest"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/muesli/goprogressbar"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestRecipients(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	act, err := newMock(ctx, u)
	assert.NoError(t, err)

	app := cli.NewApp()
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	c := cli.NewContext(app, fs, nil)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	goprogressbar.Stdout = buf
	color.NoColor = true
	defer func() {
		out.Stdout = os.Stdout
		stdout = os.Stdout
		goprogressbar.Stdout = os.Stdout
	}()

	// RecipientsPrint
	assert.NoError(t, act.RecipientsPrint(ctx, c))
	want := `Hint: run 'gopass sync' to import any missing public keys
gopass
└── 0xDEADBEEF

`
	assert.Equal(t, want, buf.String())
	buf.Reset()

	// RecipientsComplete
	act.RecipientsComplete(ctx, c)
	want = "0xDEADBEEF\n"
	assert.Equal(t, want, buf.String())
	buf.Reset()

	// RecipientsAdd
	assert.Error(t, act.RecipientsAdd(ctx, c))
	buf.Reset()

	// RecipientsRemove
	assert.Error(t, act.RecipientsRemove(ctx, c))
	buf.Reset()

	// RecipientsAdd 0xFEEDBEEF
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"0xFEEDBEEF"}))
	c = cli.NewContext(app, fs, nil)
	assert.NoError(t, act.RecipientsAdd(ctx, c))
	buf.Reset()

	// RecipientsAdd 0xBEEFFEED
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"0xBEEFFEED"}))
	c = cli.NewContext(app, fs, nil)
	assert.Error(t, act.RecipientsAdd(ctx, c))
	buf.Reset()

	// RecipientsRemove 0xDEADBEEF
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"0xDEADBEEF"}))
	c = cli.NewContext(app, fs, nil)
	assert.NoError(t, act.RecipientsRemove(ctx, c))
}
