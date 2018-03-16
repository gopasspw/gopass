package action

import (
	"bytes"
	"context"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/justwatchcom/gopass/pkg/ctxutil"
	"github.com/justwatchcom/gopass/pkg/out"
	"github.com/justwatchcom/gopass/pkg/termio"
	"github.com/justwatchcom/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestExtractHostname(t *testing.T) {
	for in, out := range map[string]string{
		"": "",
		"http://www.example.org/":              "www.example.org",
		"++#+++#jhlkadsrezu 33 553q ++++##$§&": "jhlkadsrezu_33_553q",
		"www.example.org/?foo=bar#abc":         "www.example.org",
	} {
		if got := extractHostname(in); got != out {
			t.Errorf("%s != %s", got, out)
		}
	}
}

func TestCreate(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	act, err := newMock(ctx, u)
	assert.NoError(t, err)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	app := cli.NewApp()
	// create
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	c := cli.NewContext(app, fs, nil)

	assert.Error(t, act.Create(ctx, c))
	buf.Reset()
}

func TestCreateWebsite(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithInteractive(ctx, false)
	act, err := newMock(ctx, u)
	assert.NoError(t, err)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	termio.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		stdout = os.Stdout
		termio.Stdout = os.Stdout
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

	ctx := context.Background()
	ctx = ctxutil.WithInteractive(ctx, false)
	act, err := newMock(ctx, u)
	assert.NoError(t, err)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	termio.Stdout = buf
	defer func() {
		stdout = os.Stdout
		out.Stdout = os.Stdout
		termio.Stdout = os.Stdout
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

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	act, err := newMock(ctx, u)
	assert.NoError(t, err)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	termio.Stdout = buf
	defer func() {
		stdout = os.Stdout
		out.Stdout = os.Stdout
		termio.Stdout = os.Stdout
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

func TestCreateAWS(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)
	act, err := newMock(ctx, u)
	assert.NoError(t, err)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	termio.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		stdout = os.Stdout
		termio.Stdout = os.Stdout
	}()

	// provide values on redirected stdin
	input := `account
user
ACCESSKEY
SECRETKEY
SECRETKEY

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

	assert.NoError(t, act.createAWS(ctx, c))
	buf.Reset()
}

func TestCreateGCP(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	act, err := newMock(ctx, u)
	assert.NoError(t, err)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	termio.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		stdout = os.Stdout
		termio.Stdout = os.Stdout
	}()

	tf := filepath.Join(u.Dir, "service-account.json")
	assert.NoError(t, ioutil.WriteFile(tf, []byte(`{"client_email": "foobar@example.org"}`), 0600))
	// provide values on redirected stdin
	input := tf
	input += `

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

	assert.NoError(t, act.createGCP(ctx, c))
	buf.Reset()
}
