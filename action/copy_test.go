package action

import (
	"bytes"
	"context"
	"flag"
	"io/ioutil"
	"os"
	"testing"

	"github.com/justwatchcom/gopass/utils/out"
	"github.com/urfave/cli"
)

func TestCopy(t *testing.T) {
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
	// copy foo bar
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	if err := fs.Parse([]string{"foo", "bar"}); err != nil {
		t.Fatalf("Error: %s", err)
	}
	c := cli.NewContext(app, fs, nil)

	if err := act.Copy(ctx, c); err != nil {
		t.Errorf("Error: %s", err)
	}
	buf.Reset()

	// copy not-found still-not-there
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	if err := fs.Parse([]string{"not-found", "still-not-there"}); err != nil {
		t.Fatalf("Error: %s", err)
	}
	c = cli.NewContext(app, fs, nil)

	if err := act.Copy(ctx, c); err == nil {
		t.Errorf("Should fail")
	}
	buf.Reset()

	// copy
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	c = cli.NewContext(app, fs, nil)

	if err := act.Copy(ctx, c); err == nil {
		t.Errorf("Should fail")
	}
	buf.Reset()
}
