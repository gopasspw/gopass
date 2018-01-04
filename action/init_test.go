package action

import (
	"bytes"
	"context"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestInit(t *testing.T) {
	td, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithInteractive(ctx, false)
	act, err := newMock(ctx, td)
	assert.NoError(t, err)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	app := cli.NewApp()
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	c := cli.NewContext(app, fs, nil)

	assert.NoError(t, act.Initialized(ctx, c))
	assert.Error(t, act.Init(ctx, c))
	assert.Error(t, act.InitOnboarding(ctx, c))
	assert.Equal(t, true, act.initHasUseablePrivateKeys(ctx))
	assert.Error(t, act.initCreatePrivateKey(ctx, "foo bar", "foo.bar@example.org"))

	// un-initialize the store
	assert.NoError(t, os.Remove(filepath.Join(td, "store", ".gpg-id")))
	assert.Error(t, act.Initialized(ctx, c))
}
