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
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestGenerate(t *testing.T) {
	td, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	act, err := newMock(ctx, td)
	assert.NoError(t, err)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	app := cli.NewApp()

	// generate foobar
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"foobar"}))
	c := cli.NewContext(app, fs, nil)

	assert.NoError(t, act.Generate(ctx, c))
	buf.Reset()

	// generate foobar
	// should succeed because of always yes
	assert.NoError(t, act.Generate(ctx, c))
	buf.Reset()

	// generate --force foobar
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	bf := cli.BoolFlag{
		Name:  "force",
		Usage: "force",
	}
	assert.NoError(t, bf.ApplyWithError(fs))
	assert.NoError(t, fs.Parse([]string{"--force=true", "foobar"}))
	c = cli.NewContext(app, fs, nil)

	assert.NoError(t, act.Generate(ctx, c))
	buf.Reset()

	// generate --force foobar 32
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	bf = cli.BoolFlag{
		Name:  "force",
		Usage: "force",
	}
	assert.NoError(t, bf.ApplyWithError(fs))
	assert.NoError(t, fs.Parse([]string{"--force=true", "foobar", "32"}))
	c = cli.NewContext(app, fs, nil)

	assert.NoError(t, act.Generate(ctx, c))
	buf.Reset()

	// generate --force --xkcd foobar 32
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	bf = cli.BoolFlag{
		Name:  "force",
		Usage: "force",
	}
	assert.NoError(t, bf.ApplyWithError(fs))
	bf = cli.BoolFlag{
		Name:  "xkcd",
		Usage: "xkcd",
	}
	assert.NoError(t, bf.ApplyWithError(fs))
	sf := cli.StringFlag{
		Name:  "xkcdlang",
		Usage: "xkcdlange",
		Value: "en",
	}
	assert.NoError(t, sf.ApplyWithError(fs))
	assert.NoError(t, fs.Parse([]string{"--force=true", "--xkcd=true", "foobar", "32"}))
	c = cli.NewContext(app, fs, nil)

	assert.NoError(t, act.Generate(ctx, c))
	buf.Reset()

	// generate --force --xkcd foobar baz 32
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	bf = cli.BoolFlag{
		Name:  "force",
		Usage: "force",
	}
	assert.NoError(t, bf.ApplyWithError(fs))
	bf = cli.BoolFlag{
		Name:  "xkcd",
		Usage: "xkcd",
	}
	assert.NoError(t, bf.ApplyWithError(fs))
	sf = cli.StringFlag{
		Name:  "xkcdlang",
		Usage: "xkcdlange",
		Value: "en",
	}
	assert.NoError(t, sf.ApplyWithError(fs))
	assert.NoError(t, fs.Parse([]string{"--force=true", "--xkcd=true", "foobar", "baz", "32"}))
	c = cli.NewContext(app, fs, nil)

	assert.NoError(t, act.Generate(ctx, c))
	buf.Reset()

	// generate --force --xkcd foobar baz
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	bf = cli.BoolFlag{
		Name:  "force",
		Usage: "force",
	}
	assert.NoError(t, bf.ApplyWithError(fs))
	bf = cli.BoolFlag{
		Name:  "xkcd",
		Usage: "xkcd",
	}
	assert.NoError(t, bf.ApplyWithError(fs))
	sf = cli.StringFlag{
		Name:  "xkcdlang",
		Usage: "xkcdlange",
		Value: "en",
	}
	assert.NoError(t, sf.ApplyWithError(fs))
	assert.NoError(t, fs.Parse([]string{"--force=true", "--xkcd=true", "foobar", "baz"}))
	c = cli.NewContext(app, fs, nil)

	assert.NoError(t, act.Generate(ctx, c))
	buf.Reset()

	// generate --force --xkcd --print foobar baz
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	bf = cli.BoolFlag{
		Name:  "force",
		Usage: "force",
	}
	assert.NoError(t, bf.ApplyWithError(fs))
	bf = cli.BoolFlag{
		Name:  "xkcd",
		Usage: "xkcd",
	}
	assert.NoError(t, bf.ApplyWithError(fs))
	bf = cli.BoolFlag{
		Name:  "print",
		Usage: "print",
	}
	assert.NoError(t, bf.ApplyWithError(fs))
	sf = cli.StringFlag{
		Name:  "xkcdlang",
		Usage: "xkcdlange",
		Value: "en",
	}
	assert.NoError(t, sf.ApplyWithError(fs))
	assert.NoError(t, fs.Parse([]string{"--force=true", "--print=true", "--xkcd=true", "foobar", "baz"}))
	c = cli.NewContext(app, fs, nil)

	assert.NoError(t, act.Generate(ctx, c))
	buf.Reset()
}

func TestKeyAndLength(t *testing.T) {
	app := cli.NewApp()

	for _, tc := range []struct {
		in     []string
		key    string
		length string
	}{
		{
			in:     []string{"32"},
			key:    "",
			length: "32",
		},
		{
			in:     []string{"baz"},
			key:    "baz",
			length: "",
		},
		{
			in:     []string{"baz", "32"},
			key:    "baz",
			length: "32",
		},
		{
			in:     []string{},
			key:    "",
			length: "",
		},
	} {
		fs := flag.NewFlagSet("default", flag.ContinueOnError)
		assert.NoError(t, fs.Parse(append([]string{"foobar"}, tc.in...)))
		c := cli.NewContext(app, fs, nil)
		k, l := keyAndLength(c)
		assert.Equal(t, tc.key, k, "Key from %+v", tc.in)
		assert.Equal(t, tc.length, l, "Length from %+v", tc.in)
	}
}
