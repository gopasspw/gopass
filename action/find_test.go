package action

import (
	"bytes"
	"context"
	"flag"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/justwatchcom/gopass/utils/out"
	"github.com/urfave/cli"
)

func TestFind(t *testing.T) {
	td, err := ioutil.TempDir("", "gopass-")
	if err != nil {
		t.Fatalf("Error: %s", err)
	}
	defer func() {
		_ = os.RemoveAll(td)
	}()

	ctx := context.Background()
	act, err := newMock(ctx, td)
	if err != nil {
		t.Fatalf("Error: %s", err)
	}

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	app := cli.NewApp()

	// find
	c := cli.NewContext(app, flag.NewFlagSet("default", flag.ContinueOnError), nil)
	if err := act.Find(ctx, c); err == nil || err.Error() != "Usage: action.test find arg" {
		t.Errorf("Should fail")
	}

	// find fo
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	if err := fs.Parse([]string{"fo"}); err != nil {
		t.Fatalf("Error: %s", err)
	}
	c = cli.NewContext(app, fs, nil)

	out := capture(t, func() error {
		return act.Find(ctx, c)
	})
	out = strings.TrimSpace(out)
	want := "0xDEADBEEF"
	if out != want {
		t.Errorf("'%s' != '%s'", out, want)
	}
	buf.Reset()

	// find yo
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	if err := fs.Parse([]string{"yo"}); err != nil {
		t.Fatalf("Error: %s", err)
	}
	c = cli.NewContext(app, fs, nil)

	if err := act.Find(ctx, c); err == nil {
		t.Errorf("Should fail")
	}
	buf.Reset()
}
