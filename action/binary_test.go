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
	"github.com/urfave/cli"
)

func TestBinary(t *testing.T) {
	td, err := ioutil.TempDir("", "gopass-")
	if err != nil {
		t.Fatalf("Error: %s", err)
	}
	defer func() {
		_ = os.RemoveAll(td)
	}()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = out.WithHidden(ctx, true)
	act, err := newMock(ctx, td)
	if err != nil {
		t.Fatalf("Error: %s", err)
	}

	app := cli.NewApp()

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	infile := filepath.Join(td, "input.txt")
	if err := ioutil.WriteFile(infile, []byte("0xDEADBEEF"), 0644); err != nil {
		t.Fatalf("Failed to write input file: %s", err)
	}
	if err := act.binaryCopy(ctx, infile, "bar", true); err != nil {
		t.Fatalf("Failed to move file to store: %s", err)
	}

	// no arg
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	c := cli.NewContext(app, fs, nil)
	if err := act.BinaryCat(ctx, c); err == nil {
		t.Errorf("Should fail")
	}
	if err := act.BinaryCopy(ctx, c); err == nil {
		t.Errorf("Should fail")
	}
	if err := act.BinaryMove(ctx, c); err == nil {
		t.Errorf("Should fail")
	}
	if err := act.BinarySum(ctx, c); err == nil {
		t.Errorf("Should fail")
	}

	// binary cat bar
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	if err := fs.Parse([]string{"bar"}); err != nil {
		t.Fatalf("Error: %s", err)
	}
	c = cli.NewContext(app, fs, nil)
	if err := act.BinaryCat(ctx, c); err != nil {
		t.Errorf("Should not fail")
	}

	outfile := filepath.Join(td, "output.txt")

	// binary copy bar tempdir/bar
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	if err := fs.Parse([]string{"bar", outfile}); err != nil {
		t.Fatalf("Error: %s", err)
	}
	c = cli.NewContext(app, fs, nil)
	if err := act.BinaryCopy(ctx, c); err != nil {
		t.Errorf("Should not fail")
	}

	// binary move tempdir/bar bar2
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	if err := fs.Parse([]string{outfile, "bar2"}); err != nil {
		t.Fatalf("Error: %s", err)
	}
	c = cli.NewContext(app, fs, nil)
	if err := act.BinaryMove(ctx, c); err != nil {
		t.Errorf("Should not fail")
	}

	// binary sum bar
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	if err := fs.Parse([]string{"bar"}); err != nil {
		t.Fatalf("Error: %s", err)
	}
	c = cli.NewContext(app, fs, nil)
	if err := act.BinarySum(ctx, c); err != nil {
		t.Errorf("Should not fail")
	}
}
