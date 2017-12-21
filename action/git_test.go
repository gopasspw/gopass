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

func TestGit(t *testing.T) {
	t.Skip("flaky")

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
	if err := fs.Parse([]string{"status"}); err != nil {
		t.Fatalf("Error: %s", err)
	}
	c := cli.NewContext(app, fs, nil)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	if err := act.GitInit(ctx, c); err != nil {
		t.Errorf("Error: %s", err)
	}

	out := capture(t, func() error {
		return act.Git(ctx, c)
	})
	want := `On branch master
nothing to commit, working directory clean`
	if out != want {
		t.Errorf("'%s' != '%s'", want, out)
	}
}
