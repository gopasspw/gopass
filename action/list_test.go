package action

import (
	"bytes"
	"context"
	"flag"
	"io/ioutil"
	"os"
	"testing"

	"github.com/justwatchcom/gopass/store/secret"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/urfave/cli"
)

func TestList(t *testing.T) {
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

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	app := cli.NewApp()
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	c := cli.NewContext(app, fs, nil)

	out := capture(t, func() error {
		return act.List(ctx, c)
	})
	want := `gopass
└── foo`
	if out != want {
		t.Errorf("'%s' != '%s'", out, want)
	}
	buf.Reset()

	// add foo/bar and list folder foo
	if err := act.Store.Set(ctx, "foo/bar", secret.New("123", "---\nbar: zab")); err != nil {
		t.Errorf("Failed to add secret: %s", err)
	}
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	if err := fs.Parse([]string{"foo"}); err != nil {
		t.Fatalf("Error: %s", err)
	}
	c = cli.NewContext(app, fs, nil)

	out = capture(t, func() error {
		return act.List(ctx, c)
	})
	want = `foo
└── bar`
	if out != want {
		t.Errorf("'%s' != '%s'", out, want)
		t.Logf("Out: %s", buf.String())
	}
	buf.Reset()
}
