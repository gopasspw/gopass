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

	"github.com/gopasspw/gopass/pkg/backend/crypto/xc"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/stretchr/testify/assert"
	"gopkg.in/urfave/cli.v1"
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

	c := cli.NewContext(app, fs, nil)
	assert.Error(t, GenerateKeypair(ctx, c))

	assert.NoError(t, fs.Parse([]string{"--name=foo"}))
	c = cli.NewContext(app, fs, nil)
	assert.Error(t, GenerateKeypair(ctx, c))

	assert.NoError(t, fs.Parse([]string{"--name=foo", "--email=bar"}))
	c = cli.NewContext(app, fs, nil)
	assert.Error(t, GenerateKeypair(ctx, c))

	assert.NoError(t, fs.Parse([]string{"--name=foo", "--email=bar", "--passphrase=foobar"}))
	c = cli.NewContext(app, fs, nil)
	assert.NoError(t, GenerateKeypair(ctx, c))
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

	cr, err := xc.New(td, &fakeAgent{"foobar"})
	assert.NoError(t, err)
	crypto = cr

	assert.NoError(t, crypto.CreatePrivateKeyBatch(ctx, "foobar", "foo.bar@example.org", "foobar"))

	app := cli.NewApp()
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	ff := cli.StringFlag{
		Name:  "file",
		Usage: "file",
	}
	assert.NoError(t, ff.ApplyWithError(fs))
	assert.NoError(t, fs.Parse([]string{"--file=" + plain}))
	c := cli.NewContext(app, fs, nil)
	assert.NoError(t, EncryptFile(ctx, c))

	assert.NoError(t, os.Remove(plain))

	assert.NoError(t, fs.Parse([]string{"--file=" + plain + ".xc"}))
	c = cli.NewContext(app, fs, nil)
	assert.NoError(t, DecryptFile(ctx, c))

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

	cr, err := xc.New(td, &fakeAgent{"foobar"})
	assert.NoError(t, err)
	crypto = cr

	assert.NoError(t, crypto.CreatePrivateKeyBatch(ctx, "foobar", "foo.bar@example.org", "foobar"))

	app := cli.NewApp()
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	ff := cli.StringFlag{
		Name:  "file",
		Usage: "file",
	}
	assert.NoError(t, ff.ApplyWithError(fs))
	assert.NoError(t, fs.Parse([]string{"--file=" + plain}))
	c := cli.NewContext(app, fs, nil)
	assert.NoError(t, EncryptFileStream(ctx, c))

	assert.NoError(t, os.Remove(plain))

	assert.NoError(t, fs.Parse([]string{"--file=" + plain + ".xc"}))
	c = cli.NewContext(app, fs, nil)
	assert.NoError(t, DecryptFileStream(ctx, c))

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
