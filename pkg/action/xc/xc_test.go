package xc

import (
	"bytes"
	"context"
	"flag"
	"os"
	"testing"

	"github.com/justwatchcom/gopass/pkg/ctxutil"
	"github.com/justwatchcom/gopass/pkg/out"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestListPrivateKeys(t *testing.T) {
	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	app := cli.NewApp()
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	c := cli.NewContext(app, fs, nil)
	assert.NoError(t, ListPrivateKeys(ctx, c))
}

func TestListPublicKeys(t *testing.T) {
	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	app := cli.NewApp()
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	c := cli.NewContext(app, fs, nil)
	assert.NoError(t, ListPublicKeys(ctx, c))
}

func TestGenerateKeypair(t *testing.T) {
	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

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
	defer func() {
		out.Stdout = os.Stdout
	}()

	assert.NoError(t, GenerateKeypair(ctx, c))
}

func TestXCExportPublicKey(t *testing.T) {
	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

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
	defer func() {
		out.Stdout = os.Stdout
	}()

	assert.Error(t, ExportPublicKey(ctx, c))
}

func TestImportPublicKey(t *testing.T) {
	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

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
	defer func() {
		out.Stdout = os.Stdout
	}()

	assert.Error(t, ImportPublicKey(ctx, c))
}

func TestExportPrivateKey(t *testing.T) {
	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

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
	defer func() {
		out.Stdout = os.Stdout
	}()

	assert.Error(t, ExportPrivateKey(ctx, c))
}

func TestImportPrivateKey(t *testing.T) {
	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

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
	defer func() {
		out.Stdout = os.Stdout
	}()

	assert.Error(t, ImportPrivateKey(ctx, c))
}
