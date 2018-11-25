package binary

import (
	"bytes"
	"context"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/tests/mockstore"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli"
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

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	infile := filepath.Join(tempdir, "input.txt")
	assert.NoError(t, ioutil.WriteFile(infile, []byte("0xDEADBEEF"), 0644))
	assert.NoError(t, binaryCopy(ctx, c, infile, "bar", true, store))

	assert.Error(t, Cat(ctx, c, store))
	assert.Error(t, Copy(ctx, c, store))
	assert.Error(t, Move(ctx, c, store))
	assert.Error(t, Sum(ctx, c, store))
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
	assert.NoError(t, Cat(ctx, c, store))
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
	assert.NoError(t, Copy(ctx, c, store))

	// binary copy tempdir/bar tempdir/bar
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{outfile, outfile}))
	c = cli.NewContext(app, fs, nil)
	assert.Error(t, Copy(ctx, c, store))

	// binary copy bar bar
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"bar", "bar"}))
	c = cli.NewContext(app, fs, nil)
	assert.Error(t, Copy(ctx, c, store))

	// binary move tempdir/bar bar2
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{outfile, "bar2"}))
	c = cli.NewContext(app, fs, nil)
	assert.NoError(t, Move(ctx, c, store))
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
	assert.NoError(t, Sum(ctx, c, store))
}
