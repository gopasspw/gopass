package action

import (
	"bytes"
	"context"
	"flag"
	"os"
	"testing"

	"github.com/justwatchcom/gopass/tests/gptest"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
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
	defer func() {
		out.Stdout = os.Stdout
		stdout = os.Stdout
	}()

	// RecipientsPrint
	assert.NoError(t, act.RecipientsPrint(ctx, c))
	want := `Hint: run 'gopass sync' to import any missing public keys
gopass
└── 0xDEADBEEF (missing public key)

`
	assert.Equal(t, want, buf.String())
	buf.Reset()

	// RecipientsComplete
	act.RecipientsComplete(ctx, c)
	want = "0xDEADBEEF (missing public key)\n"
	assert.Equal(t, want, buf.String())
	buf.Reset()

	// RecipientsAdd
	assert.Error(t, act.RecipientsAdd(ctx, c))
	buf.Reset()

	// RecipientsRemove
	assert.Error(t, act.RecipientsRemove(ctx, c))
	buf.Reset()
}
