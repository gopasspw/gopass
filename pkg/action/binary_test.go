package action

import (
	"bytes"
	"context"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/justwatchcom/gopass/pkg/ctxutil"
	"github.com/justwatchcom/gopass/pkg/out"
	"github.com/justwatchcom/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestBinary(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = out.WithHidden(ctx, true)
	act, err := newMock(ctx, u)
	assert.NoError(t, err)

	app := cli.NewApp()

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	infile := filepath.Join(u.Dir, "input.txt")
	assert.NoError(t, ioutil.WriteFile(infile, []byte("0xDEADBEEF"), 0644))
	assert.NoError(t, act.binaryCopy(ctx, infile, "bar", true))

	// no arg
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	c := cli.NewContext(app, fs, nil)
	assert.Error(t, act.BinaryCat(ctx, c))
	assert.Error(t, act.BinaryCopy(ctx, c))
	assert.Error(t, act.BinaryMove(ctx, c))
	assert.Error(t, act.BinarySum(ctx, c))
}

func TestBinaryCat(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = out.WithHidden(ctx, true)
	act, err := newMock(ctx, u)
	assert.NoError(t, err)

	app := cli.NewApp()

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	infile := filepath.Join(u.Dir, "input.txt")
	assert.NoError(t, ioutil.WriteFile(infile, []byte("0xDEADBEEF"), 0644))
	assert.NoError(t, act.binaryCopy(ctx, infile, "bar", true))

	// binary cat bar
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"bar"}))
	c := cli.NewContext(app, fs, nil)
	assert.NoError(t, act.BinaryCat(ctx, c))
}

func TestBinaryCopy(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = out.WithHidden(ctx, true)
	act, err := newMock(ctx, u)
	assert.NoError(t, err)

	app := cli.NewApp()

	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		stdout = os.Stdout
	}()

	infile := filepath.Join(u.Dir, "input.txt")
	assert.NoError(t, ioutil.WriteFile(infile, []byte("0xDEADBEEF"), 0644))
	assert.NoError(t, act.binaryCopy(ctx, infile, "bar", true))

	outfile := filepath.Join(u.Dir, "output.txt")

	// binary copy bar tempdir/bar
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"bar", outfile}))
	c := cli.NewContext(app, fs, nil)
	assert.NoError(t, act.BinaryCopy(ctx, c))

	// binary copy tempdir/bar tempdir/bar
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{outfile, outfile}))
	c = cli.NewContext(app, fs, nil)
	assert.Error(t, act.BinaryCopy(ctx, c))

	// binary copy bar bar
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"bar", "bar"}))
	c = cli.NewContext(app, fs, nil)
	assert.Error(t, act.BinaryCopy(ctx, c))

	// binary move tempdir/bar bar2
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{outfile, "bar2"}))
	c = cli.NewContext(app, fs, nil)
	assert.NoError(t, act.BinaryMove(ctx, c))
}

func TestBinarySum(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = out.WithHidden(ctx, true)
	act, err := newMock(ctx, u)
	assert.NoError(t, err)

	app := cli.NewApp()

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	infile := filepath.Join(u.Dir, "input.txt")
	assert.NoError(t, ioutil.WriteFile(infile, []byte("0xDEADBEEF"), 0644))
	assert.NoError(t, act.binaryCopy(ctx, infile, "bar", true))

	// binary sum bar
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"bar"}))
	c := cli.NewContext(app, fs, nil)
	assert.NoError(t, act.BinarySum(ctx, c))
}
