package action

import (
	"bytes"
	"context"
	"flag"
	"os"
	"testing"

	"github.com/justwatchcom/gopass/tests/gptest"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestXCListPrivateKeys(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	act, err := newMock(ctx, u)
	assert.NoError(t, err)

	app := cli.NewApp()
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	c := cli.NewContext(app, fs, nil)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		stdout = os.Stdout
	}()

	assert.NoError(t, act.XCListPrivateKeys(ctx, c))
}

func TestXCListPublicKeys(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	act, err := newMock(ctx, u)
	assert.NoError(t, err)

	app := cli.NewApp()
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	c := cli.NewContext(app, fs, nil)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		stdout = os.Stdout
	}()

	assert.NoError(t, act.XCListPublicKeys(ctx, c))
}

func TestXCGenerateKeypair(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	act, err := newMock(ctx, u)
	assert.NoError(t, err)

	app := cli.NewApp()
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	nf := cli.StringFlag{
		Name:  "name",
		Usage: "name",
	}
	assert.NoError(t, nf.ApplyWithError(fs))
	ef := cli.StringFlag{
		Name:  "email",
		Usage: "email",
	}
	assert.NoError(t, ef.ApplyWithError(fs))
	pf := cli.StringFlag{
		Name:  "passphrase",
		Usage: "passphrase",
	}
	assert.NoError(t, pf.ApplyWithError(fs))
	assert.NoError(t, fs.Parse([]string{"--name=foo", "--email=bar", "--passphrase=foobar"}))
	c := cli.NewContext(app, fs, nil)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		stdout = os.Stdout
	}()

	assert.NoError(t, act.XCGenerateKeypair(ctx, c))
}

func TestXCExportPublicKey(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	act, err := newMock(ctx, u)
	assert.NoError(t, err)

	app := cli.NewApp()
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	id := cli.StringFlag{
		Name:  "id",
		Usage: "id",
	}
	assert.NoError(t, id.ApplyWithError(fs))
	ff := cli.StringFlag{
		Name:  "file",
		Usage: "file",
	}
	assert.NoError(t, ff.ApplyWithError(fs))
	assert.NoError(t, fs.Parse([]string{"--id=foo", "--file=/tmp/foo.pub"}))
	c := cli.NewContext(app, fs, nil)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		stdout = os.Stdout
	}()

	assert.Error(t, act.XCExportPublicKey(ctx, c))
}

func TestXCImportPublicKey(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	act, err := newMock(ctx, u)
	assert.NoError(t, err)

	app := cli.NewApp()
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	ff := cli.StringFlag{
		Name:  "file",
		Usage: "file",
	}
	assert.NoError(t, ff.ApplyWithError(fs))
	assert.NoError(t, fs.Parse([]string{"--file=/tmp/foo.pub"}))
	c := cli.NewContext(app, fs, nil)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		stdout = os.Stdout
	}()

	assert.Error(t, act.XCImportPublicKey(ctx, c))
}

func TestXCExportPrivateKey(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	act, err := newMock(ctx, u)
	assert.NoError(t, err)

	app := cli.NewApp()
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	id := cli.StringFlag{
		Name:  "id",
		Usage: "id",
	}
	assert.NoError(t, id.ApplyWithError(fs))
	ff := cli.StringFlag{
		Name:  "file",
		Usage: "file",
	}
	assert.NoError(t, ff.ApplyWithError(fs))
	assert.NoError(t, fs.Parse([]string{"--id=foo", "--file=/tmp/foo.pub"}))
	c := cli.NewContext(app, fs, nil)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		stdout = os.Stdout
	}()

	assert.Error(t, act.XCExportPrivateKey(ctx, c))
}

func TestXCImportPrivateKey(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	act, err := newMock(ctx, u)
	assert.NoError(t, err)

	app := cli.NewApp()
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	ff := cli.StringFlag{
		Name:  "file",
		Usage: "file",
	}
	assert.NoError(t, ff.ApplyWithError(fs))
	assert.NoError(t, fs.Parse([]string{"--file=/tmp/foo.pub"}))
	c := cli.NewContext(app, fs, nil)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		stdout = os.Stdout
	}()

	assert.Error(t, act.XCImportPrivateKey(ctx, c))
}
