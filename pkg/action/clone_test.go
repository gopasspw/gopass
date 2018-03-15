package action

import (
	"bytes"
	"context"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/justwatchcom/gopass/pkg/backend"
	git "github.com/justwatchcom/gopass/pkg/backend/rcs/git/cli"
	"github.com/justwatchcom/gopass/pkg/ctxutil"
	"github.com/justwatchcom/gopass/pkg/out"
	"github.com/justwatchcom/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestClone(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithInteractive(ctx, false)
	ctx = backend.WithRCSBackend(ctx, backend.GitCLI)

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

	// no args
	assert.Error(t, act.Clone(ctx, c))

	// clone to initialized store
	assert.Error(t, act.clone(ctx, "/tmp/non-existing-repo.git", "", filepath.Join(u.Dir, "store")))

	// clone to mount
	gd := filepath.Join(u.Dir, "other-repo")
	assert.NoError(t, os.MkdirAll(gd, 0700))
	_, err = git.Open(gd, "")
	assert.Error(t, err)
	idf := filepath.Join(gd, ".gpg-id")
	assert.NoError(t, ioutil.WriteFile(idf, []byte("0xDEADBEEF"), 0600))
	gr, err := git.Init(ctx, gd, "Nobody", "foo.bar@example.org")
	assert.NoError(t, err)
	assert.NotNil(t, gr)
	assert.NoError(t, act.clone(ctx, gd, "gd", filepath.Join(u.Dir, "mount")))
}

func TestCloneGetGitConfig(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithInteractive(ctx, false)

	act, err := newMock(ctx, u)
	assert.NoError(t, err)

	name, email, err := act.cloneGetGitConfig(ctx, "foobar")
	assert.NoError(t, err)
	assert.Equal(t, "", name)
	assert.Equal(t, "", email)
}

func TestDetectCryptoBackend(t *testing.T) {
	ctx := context.Background()

	tempdir, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	gpgdir := filepath.Join(tempdir, ".password-store-gpg")
	gpgfn := filepath.Join(gpgdir, ".gpg-id")
	assert.NoError(t, os.Mkdir(gpgdir, 0755))
	assert.NoError(t, ioutil.WriteFile(gpgfn, []byte("foobar"), 0644))

	xcdir := filepath.Join(tempdir, ".password-store-xc")
	xcfn := filepath.Join(xcdir, ".xc-ids")
	assert.NoError(t, os.Mkdir(xcdir, 0755))
	assert.NoError(t, ioutil.WriteFile(xcfn, []byte("foobar"), 0644))

	assert.Equal(t, backend.GPGCLI, detectCryptoBackend(ctx, gpgdir))
	assert.Equal(t, backend.XC, detectCryptoBackend(ctx, xcdir))
}
