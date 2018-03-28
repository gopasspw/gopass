package create

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
	"github.com/justwatchcom/gopass/tests/mockstore"
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
	store := mockstore.New("")

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	app := cli.NewApp()
	// create
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	c := cli.NewContext(app, fs, nil)

	assert.Error(t, Create(ctx, c, store))
	buf.Reset()
}

func TestCreateWebsite(t *testing.T) {
	s := creator{mockstore.New("")}

	ctx := context.Background()
	ctx = ctxutil.WithInteractive(ctx, true)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	termio.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
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

	assert.NoError(t, s.createWebsite(ctx, c))
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

	assert.NoError(t, s.createWebsite(ctx, c))
	buf.Reset()
}

func TestCreatePIN(t *testing.T) {
	s := creator{mockstore.New("")}

	ctx := context.Background()
	ctx = ctxutil.WithInteractive(ctx, true)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	termio.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		termio.Stdout = os.Stdout
	}()

	ctx = ctxutil.WithAlwaysYes(ctx, true)
	pw, err := s.createGeneratePIN(ctx)
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

	assert.NoError(t, s.createPIN(ctx, c))
	buf.Reset()
}

func TestCreateGeneric(t *testing.T) {
	s := creator{mockstore.New("")}

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	termio.Stdout = buf
	defer func() {
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

	assert.NoError(t, s.createGeneric(ctx, c))
	buf.Reset()
}

func TestCreateAWS(t *testing.T) {
	s := creator{mockstore.New("")}

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	termio.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
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

	assert.NoError(t, s.createAWS(ctx, c))
	buf.Reset()
}

func TestCreateGCP(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	s := creator{mockstore.New("")}

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	termio.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		termio.Stdout = os.Stdout
	}()

	tf := filepath.Join(tempdir, "service-account.json")
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

	assert.NoError(t, s.createGCP(ctx, c))
	buf.Reset()
}
