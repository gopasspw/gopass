package binary

import (
	"bytes"
	"context"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/mockstore"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestBinary(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "gopass-")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	store := mockstore.New("")

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = out.WithHidden(ctx, true)

	app := cli.NewApp()
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	c := cli.NewContext(app, fs, nil)
	c.Context = ctx

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	infile := filepath.Join(tempdir, "input.txt")
	assert.NoError(t, ioutil.WriteFile(infile, []byte("0xDEADBEEF"), 0644))
	assert.NoError(t, binaryCopy(ctx, c, infile, "bar", true, store))

	assert.Error(t, Cat(c, store))
	assert.Error(t, Copy(c, store))
	assert.Error(t, Move(c, store))
	assert.Error(t, Sum(c, store))
}

func TestBinaryCat(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "gopass-")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	store := mockstore.New("")

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = out.WithHidden(ctx, true)

	app := cli.NewApp()
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	c := cli.NewContext(app, fs, nil)
	c.Context = ctx

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	infile := filepath.Join(tempdir, "input.txt")
	assert.NoError(t, ioutil.WriteFile(infile, []byte("0xDEADBEEF"), 0644))
	assert.NoError(t, binaryCopy(ctx, c, infile, "bar", true, store))

	// binary cat bar
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"bar"}))
	c = cli.NewContext(app, fs, nil)
	assert.NoError(t, Cat(c, store))
}

func TestBinaryCopy(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "gopass-")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	store := mockstore.New("")

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = out.WithHidden(ctx, true)

	app := cli.NewApp()
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	c := cli.NewContext(app, fs, nil)
	c.Context = ctx

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	infile := filepath.Join(tempdir, "input.txt")
	assert.NoError(t, ioutil.WriteFile(infile, []byte("0xDEADBEEF"), 0644))
	assert.NoError(t, binaryCopy(ctx, c, infile, "bar", true, store))

	outfile := filepath.Join(tempdir, "output.txt")

	// binary copy bar tempdir/bar
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"bar", outfile}))
	c = cli.NewContext(app, fs, nil)
	c.Context = ctx

	assert.NoError(t, Copy(c, store))

	// binary copy tempdir/bar tempdir/bar
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{outfile, outfile}))
	c = cli.NewContext(app, fs, nil)
	c.Context = ctx

	assert.Error(t, Copy(c, store))

	// binary copy bar bar
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"bar", "bar"}))
	c = cli.NewContext(app, fs, nil)
	c.Context = ctx

	assert.Error(t, Copy(c, store))

	// binary move tempdir/bar bar2
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{outfile, "bar2"}))
	c = cli.NewContext(app, fs, nil)
	c.Context = ctx

	assert.NoError(t, Move(c, store))
}

func TestBinarySum(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "gopass-")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	store := mockstore.New("")

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = out.WithHidden(ctx, true)

	app := cli.NewApp()
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	c := cli.NewContext(app, fs, nil)
	c.Context = ctx

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	infile := filepath.Join(tempdir, "input.txt")
	assert.NoError(t, ioutil.WriteFile(infile, []byte("0xDEADBEEF"), 0644))
	assert.NoError(t, binaryCopy(ctx, c, infile, "bar", true, store))

	// binary sum bar
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"bar"}))
	c = cli.NewContext(app, fs, nil)
	c.Context = ctx

	assert.NoError(t, Sum(c, store))
}
