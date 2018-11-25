package action

import (
	"bytes"
	"context"
	"flag"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/tests/gptest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli"
)

func TestGenerate(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithAutoClip(ctx, false)

	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()
	color.NoColor = true

	app := cli.NewApp()

	// generate
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	c := cli.NewContext(app, fs, nil)

	assert.Error(t, act.Generate(ctx, c))
	buf.Reset()

	// generate foobar
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"foobar"}))
	c = cli.NewContext(app, fs, nil)

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

	// generate --force --symbols foobar 32
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	bf = cli.BoolFlag{
		Name:  "force",
		Usage: "force",
	}
	assert.NoError(t, bf.ApplyWithError(fs))
	bf = cli.BoolFlag{
		Name:  "symbols",
		Usage: "symbols",
	}
	assert.NoError(t, bf.ApplyWithError(fs))
	bf = cli.BoolFlag{
		Name:  "print",
		Usage: "print",
	}
	assert.NoError(t, bf.ApplyWithError(fs))
	assert.NoError(t, fs.Parse([]string{"--force=true", "--print=true", "--symbols", "foobar", "32"}))
	c = cli.NewContext(app, fs, nil)

	assert.NoError(t, act.Generate(ctx, c))
	passIsAlphaNum(t, buf.String(), false)
	buf.Reset()

	// generate --force --symbols=true foobar 32
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	bf = cli.BoolFlag{
		Name:  "force",
		Usage: "force",
	}
	assert.NoError(t, bf.ApplyWithError(fs))
	bf = cli.BoolFlag{
		Name:  "symbols",
		Usage: "symbols",
	}
	assert.NoError(t, bf.ApplyWithError(fs))
	bf = cli.BoolFlag{
		Name:  "print",
		Usage: "print",
	}
	assert.NoError(t, bf.ApplyWithError(fs))
	assert.NoError(t, fs.Parse([]string{"--force=true", "--print=true", "--symbols=true", "foobar", "32"}))
	c = cli.NewContext(app, fs, nil)

	assert.NoError(t, act.Generate(ctx, c))
	passIsAlphaNum(t, buf.String(), false)
	buf.Reset()

	// generate --force --symbols=false foobar 32
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	bf = cli.BoolFlag{
		Name:  "force",
		Usage: "force",
	}
	assert.NoError(t, bf.ApplyWithError(fs))
	bf = cli.BoolFlag{
		Name:  "symbols",
		Usage: "symbols",
	}
	assert.NoError(t, bf.ApplyWithError(fs))
	bf = cli.BoolFlag{
		Name:  "print",
		Usage: "print",
	}
	assert.NoError(t, bf.ApplyWithError(fs))
	assert.NoError(t, fs.Parse([]string{"--force=true", "--print=true", "--symbols=false", "foobar", "32"}))
	c = cli.NewContext(app, fs, nil)

	assert.NoError(t, act.Generate(ctx, c))
	passIsAlphaNum(t, buf.String(), true)
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

func passIsAlphaNum(t *testing.T, buf string, want bool) {
	reAlphaNum := regexp.MustCompile(`^[A-Za-z0-9]+$`)
	lines := strings.Split(strings.TrimSpace(buf), "\n")
	if len(lines) < 1 {
		t.Errorf("buffer empty (no lines)")
	}
	line := strings.TrimSpace(lines[len(lines)-1])
	if reAlphaNum.MatchString(line) != want {
		t.Errorf("buffer did not match alpha num re: %s (%s)", line, buf)
	}
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
		args, _ := parseArgs(c)
		k, l := keyAndLength(args)
		assert.Equal(t, tc.key, k, "Key from %+v", tc.in)
		assert.Equal(t, tc.length, l, "Length from %+v", tc.in)
	}
}

func TestExtractEmails(t *testing.T) {
	for _, tc := range []struct {
		in  []string
		out []string
	}{
		{
			out: []string{},
		},
		{
			in:  []string{"some/mount/gmail.com/john.doe@example.org", "example.com/user@example.org"},
			out: []string{"john.doe@example.org", "user@example.org"},
		},
	} {
		assert.Equal(t, tc.out, extractEmails(tc.in))
	}
}

func TestExtractDomains(t *testing.T) {
	for _, tc := range []struct {
		in  []string
		out []string
	}{
		{
			out: []string{},
		},
		{
			in:  []string{"websites/gmail.com", "live.com", "some/mount/websites/web.de"},
			out: []string{"gmail.com", "live.com", "web.de"},
		},
	} {
		assert.Equal(t, tc.out, extractDomains(tc.in))
	}
}

func TestUniq(t *testing.T) {
	for _, tc := range []struct {
		in  []string
		out []string
	}{
		{
			out: []string{},
		},
		{
			in:  []string{"foo", "foo", "bar"},
			out: []string{"bar", "foo"},
		},
	} {
		assert.Equal(t, tc.out, uniq(tc.in))
	}
}

func TestFilterPrefix(t *testing.T) {
	for _, tc := range []struct {
		in     []string
		prefix string
		out    []string
	}{
		{
			out: []string{},
		},
		{
			in:     []string{"foo", "bar", "baz"},
			prefix: "foo",
			out:    []string{"foo"},
		},
		{
			in:     []string{"foo/bar", "foo/baz", "bar/foo"},
			prefix: "foo/",
			out:    []string{"foo/bar", "foo/baz"},
		},
	} {
		assert.Equal(t, tc.out, filterPrefix(tc.in, tc.prefix))
	}
}
