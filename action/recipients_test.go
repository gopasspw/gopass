package action

import (
	"bytes"
	"context"
	"flag"
	"io/ioutil"
	"os"
	"testing"

	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/urfave/cli"
)

func TestRecipients(t *testing.T) {
	td, err := ioutil.TempDir("", "gopass-")
	if err != nil {
		t.Fatalf("Error: %s", err)
	}
	defer func() {
		_ = os.RemoveAll(td)
	}()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	act, err := newMock(ctx, td)
	if err != nil {
		t.Fatalf("Error: %s", err)
	}

	app := cli.NewApp()
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	c := cli.NewContext(app, fs, nil)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	// RecipientsPrint
	out := capture(t, func() error {
		return act.RecipientsPrint(ctx, c)
	})
	want := `gopass
└── 0xDEADBEEF (missing public key)`
	if out != want {
		t.Errorf("'%s' != '%s'", want, out)
	}
	buf.Reset()

	// RecipientsComplete
	out = capture(t, func() error {
		act.RecipientsComplete(ctx, c)
		return nil
	})
	want = "0xDEADBEEF (missing public key)"
	if out != want {
		t.Errorf("'%s' != '%s'", want, out)
	}
	buf.Reset()

	// RecipientsAdd
	if err := act.RecipientsAdd(ctx, c); err == nil {
		t.Errorf("Should fail")
	}
	buf.Reset()

	// RecipientsRemove
	if err := act.RecipientsRemove(ctx, c); err == nil {
		t.Errorf("Should fail")
	}
	buf.Reset()
}
