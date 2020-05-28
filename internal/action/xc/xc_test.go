// +build xc

package xc

import (
	"bytes"
	"context"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gopasspw/gopass/internal/backend/crypto/xc"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
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
	c.Context = ctx
	assert.NoError(t, ListPrivateKeys(c))
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
	c.Context = ctx
	assert.NoError(t, ListPublicKeys(c))
}

func TestGenerateKeypair(t *testing.T) {
	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	app := cli.NewApp()
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	nf := cli.StringFlag{
		Name:  "name",
		Usage: "name",
	}
	assert.NoError(t, nf.Apply(fs))
	ef := cli.StringFlag{
		Name:  "email",
		Usage: "email",
	}
	assert.NoError(t, ef.Apply(fs))
	pf := cli.StringFlag{
		Name:  "passphrase",
		Usage: "passphrase",
	}
	assert.NoError(t, pf.Apply(fs))

	c := cli.NewContext(app, fs, nil)
	c.Context = ctx
	assert.Error(t, GenerateKeypair(c))

	assert.NoError(t, fs.Parse([]string{"--name=foo"}))
	c = cli.NewContext(app, fs, nil)
	c.Context = ctx
	assert.Error(t, GenerateKeypair(c))

	assert.NoError(t, fs.Parse([]string{"--name=foo", "--email=bar"}))
	c = cli.NewContext(app, fs, nil)
	c.Context = ctx
	assert.Error(t, GenerateKeypair(c))

	assert.NoError(t, fs.Parse([]string{"--name=foo", "--email=bar", "--passphrase=foobar"}))
	c = cli.NewContext(app, fs, nil)
	c.Context = ctx
	assert.NoError(t, GenerateKeypair(c))
}

func TestExportPublicKey(t *testing.T) {
	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	app := cli.NewApp()
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	id := cli.StringFlag{
		Name:  "id",
		Usage: "id",
	}
	assert.NoError(t, id.Apply(fs))
	ff := cli.StringFlag{
		Name:  "file",
		Usage: "file",
	}
	assert.NoError(t, ff.Apply(fs))
	assert.NoError(t, fs.Parse([]string{"--id=foo", "--file=/tmp/foo.pub"}))
	c := cli.NewContext(app, fs, nil)
	c.Context = ctx

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	assert.Error(t, ExportPublicKey(c))
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
	assert.NoError(t, ff.Apply(fs))
	assert.NoError(t, fs.Parse([]string{"--file=/tmp/foo.pub"}))
	c := cli.NewContext(app, fs, nil)
	c.Context = ctx

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	assert.Error(t, ImportPublicKey(c))
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
	assert.NoError(t, id.Apply(fs))
	ff := cli.StringFlag{
		Name:  "file",
		Usage: "file",
	}
	assert.NoError(t, ff.Apply(fs))
	assert.NoError(t, fs.Parse([]string{"--id=foo", "--file=/tmp/foo.pub"}))
	c := cli.NewContext(app, fs, nil)
	c.Context = ctx

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	assert.Error(t, ExportPrivateKey(c))
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
	assert.NoError(t, ff.Apply(fs))
	assert.NoError(t, fs.Parse([]string{"--file=/tmp/foo.pub"}))
	c := cli.NewContext(app, fs, nil)
	c.Context = ctx

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	assert.Error(t, ImportPrivateKey(c))
}

func TestEncryptDecryptFile(t *testing.T) {
	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	td, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	plain := filepath.Join(td, "plain.txt")
	assert.NoError(t, ioutil.WriteFile(plain, []byte("foobar"), 0600))

	crypto = xc.New(td, &fakeAgent{"foobar"})

	assert.NoError(t, crypto.CreatePrivateKeyBatch(ctx, "foobar", "foo.bar@example.org", "foobar"))

	app := cli.NewApp()
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	ff := cli.StringFlag{
		Name:  "file",
		Usage: "file",
	}
	assert.NoError(t, ff.Apply(fs))
	assert.NoError(t, fs.Parse([]string{"--file=" + plain}))
	c := cli.NewContext(app, fs, nil)
	c.Context = ctx

	assert.NoError(t, EncryptFile(c))
	assert.NoError(t, os.Remove(plain))

	assert.NoError(t, fs.Parse([]string{"--file=" + plain + ".xc"}))

	c = cli.NewContext(app, fs, nil)
	c.Context = ctx
	assert.NoError(t, DecryptFile(c))

	content, err := ioutil.ReadFile(plain)
	assert.NoError(t, err)
	assert.Equal(t, "foobar", string(content))
}

func TestEncryptDecryptStream(t *testing.T) {
	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	td, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	plain := filepath.Join(td, "plain.txt")
	assert.NoError(t, ioutil.WriteFile(plain, []byte("foobar"), 0600))

	crypto = xc.New(td, &fakeAgent{"foobar"})

	assert.NoError(t, crypto.CreatePrivateKeyBatch(ctx, "foobar", "foo.bar@example.org", "foobar"))

	app := cli.NewApp()
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	ff := cli.StringFlag{
		Name:  "file",
		Usage: "file",
	}
	assert.NoError(t, ff.Apply(fs))
	assert.NoError(t, fs.Parse([]string{"--file=" + plain}))

	c := cli.NewContext(app, fs, nil)
	c.Context = ctx

	assert.NoError(t, EncryptFileStream(c))
	assert.NoError(t, os.Remove(plain))

	assert.NoError(t, fs.Parse([]string{"--file=" + plain + ".xc"}))
	c = cli.NewContext(app, fs, nil)
	c.Context = ctx

	assert.NoError(t, DecryptFileStream(c))

	content, err := ioutil.ReadFile(plain)
	assert.NoError(t, err)
	assert.Equal(t, "foobar", strings.TrimSpace(string(content)))
}

type fakeAgent struct {
	pw string
}

func (f *fakeAgent) Ping(context.Context) error {
	return nil
}

func (f *fakeAgent) Remove(context.Context, string) error {
	return nil
}

func (f *fakeAgent) Passphrase(context.Context, string, string) (string, error) {
	return f.pw, nil
}
