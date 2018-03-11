package action

import (
	"bytes"
	"context"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/justwatchcom/gopass/tests/gptest"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestInit(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithInteractive(ctx, false)
	ctx = ctxutil.WithDebug(ctx, true)
	act, err := newMock(ctx, u)
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
	crypto := act.Store.Crypto(ctx, "")
	assert.Equal(t, true, act.initHasUseablePrivateKeys(ctx, crypto, ""))
	assert.Error(t, act.initCreatePrivateKey(ctx, crypto, "", "foo bar", "foo.bar@example.org"))
	buf.Reset()

	// un-initialize the store
	assert.NoError(t, os.Remove(filepath.Join(u.StoreDir(""), ".gpg-id")))
	assert.Error(t, act.Initialized(ctx, c))
	buf.Reset()
}
