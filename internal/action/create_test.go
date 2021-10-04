package action

import (
	"bytes"
	"context"
	"flag"
	"os"
	"strings"
	"testing"

	aclip "github.com/atotto/clipboard"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/termio"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestExtractHostname(t *testing.T) {
	for in, out := range map[string]string{
		"":                                     "",
		"http://www.example.org/":              "www.example.org",
		"++#+++#jhlkadsrezu 33 553q ++++##$§&": "jhlkadsrezu_33_553q",
		"www.example.org/?foo=bar#abc":         "www.example.org",
		"a test":                               "a_test",
	} {
		assert.Equal(t, out, extractHostname(in))
	}
}

func TestCreate(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	aclip.Unsupported = true

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithNotifications(ctx, false)

	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	act.cfg.ClipTimeout = 1

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	// create
	c := gptest.CliCtx(ctx, t)

	assert.Error(t, act.Create(c))
	buf.Reset()
}

func TestCreateWebsite(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	aclip.Unsupported = true

	ctx := context.Background()
	ctx = ctxutil.WithInteractive(ctx, true)
	ctx = ctxutil.WithNotifications(ctx, false)

	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	act.cfg.ClipTimeout = 1

	buf := &bytes.Buffer{}
	out.Stderr = buf
	termio.Stderr = buf
	defer func() {
		out.Stderr = os.Stderr
		termio.Stderr = os.Stderr
	}()

	// provide values on redirected stdin
	input := `https://www.example.org/
foobar
y
y
5
`
	termio.Stdin = strings.NewReader(input)
	ctx = ctxutil.WithAlwaysYes(ctx, false)
	defer func() {
		termio.Stdin = os.Stdin
	}()

	app := cli.NewApp()
	// create
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	sf := cli.BoolFlag{
		Name:  "print",
		Usage: "print",
	}
	assert.NoError(t, sf.Apply(fs))
	assert.NoError(t, fs.Parse([]string{"--print=true"}))
	c := cli.NewContext(app, fs, nil)

	assert.NoError(t, act.createWebsite(ctx, c))
	buf.Reset()

	// try to create the same entry twice
	input = `https://www.example.org/
foobar
y
y
5
`
	termio.Stdin = strings.NewReader(input)

	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	c = cli.NewContext(app, fs, nil)

	assert.NoError(t, act.createWebsite(ctx, c))
	buf.Reset()
}

func TestCreatePIN(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	aclip.Unsupported = true

	ctx := context.Background()
	ctx = ctxutil.WithInteractive(ctx, true)
	ctx = ctxutil.WithNotifications(ctx, false)

	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	act.cfg.ClipTimeout = 1

	buf := &bytes.Buffer{}
	out.Stderr = buf
	termio.Stderr = buf
	defer func() {
		out.Stderr = os.Stderr
		termio.Stderr = os.Stderr
	}()

	ctx = ctxutil.WithAlwaysYes(ctx, true)

	pw, err := act.createGeneratePIN(ctx)
	assert.NoError(t, err)
	if len(pw) < 4 || len(pw) > 4 {
		t.Errorf("PIN should have 4 characters")
	}
	buf.Reset()

	// provide values on redirected stdin
	input := `MyBank
FooCard
y
8
`
	termio.Stdin = strings.NewReader(input)
	ctx = ctxutil.WithAlwaysYes(ctx, false)
	defer func() {
		termio.Stdin = os.Stdin
	}()

	app := cli.NewApp()
	// create
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	c := cli.NewContext(app, fs, nil)

	assert.NoError(t, act.createPIN(ctx, c))
	buf.Reset()
}

func TestCreateGeneric(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	aclip.Unsupported = true

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithNotifications(ctx, false)

	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	act.cfg.ClipTimeout = 1

	buf := &bytes.Buffer{}
	out.Stderr = buf
	termio.Stderr = buf
	defer func() {
		out.Stderr = os.Stderr
		termio.Stderr = os.Stderr
	}()

	// provide values on redirected stdin
	input := `foobar
y
y
8

`
	termio.Stdin = strings.NewReader(input)
	ctx = ctxutil.WithAlwaysYes(ctx, false)
	defer func() {
		termio.Stdin = os.Stdin
	}()

	app := cli.NewApp()
	// create
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	c := cli.NewContext(app, fs, nil)

	assert.NoError(t, act.createGeneric(ctx, c))
	buf.Reset()
}
