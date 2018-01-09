package action

import (
	"bytes"
	"context"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	git "github.com/justwatchcom/gopass/backend/git/cli"
	"github.com/justwatchcom/gopass/tests/gptest"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestClone(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithInteractive(ctx, false)

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

	t.Skip("flaky")

	// clone to mount
	gd := filepath.Join(u.Dir, "other-repo")
	assert.NoError(t, os.MkdirAll(gd, 0700))
	gr := git.New(gd, "")
	idf := filepath.Join(gd, ".gpg-id")
	assert.NoError(t, ioutil.WriteFile(idf, []byte("0xDEADBEEF"), 0600))
	assert.NoError(t, gr.Init(ctx, "", "", ""))
	assert.NoError(t, act.clone(ctx, gd, "gd", filepath.Join(u.Dir, "mount")))
}
