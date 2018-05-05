package action

import (
	"bytes"
	"context"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/justwatchcom/gopass/pkg/ctxutil"
	"github.com/justwatchcom/gopass/pkg/out"
	"github.com/justwatchcom/gopass/tests/gptest"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestMounts(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	act, err := newMock(ctx, u)
	assert.NoError(t, err)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		stdout = os.Stdout
	}()

	app := cli.NewApp()

	// print mounts
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	c := cli.NewContext(app, fs, nil)

	assert.NoError(t, act.MountsPrint(ctx, c))
	buf.Reset()

	// complete mounts
	act.MountsComplete(c)
	if buf.String() != "" {
		t.Errorf("Should be empty")
	}
	buf.Reset()

	// remove no non-existing mount
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	c = cli.NewContext(app, fs, nil)

	assert.Error(t, act.MountRemove(ctx, c))
	buf.Reset()

	// remove non-existing mount
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"foo"}))
	c = cli.NewContext(app, fs, nil)

	assert.NoError(t, act.MountRemove(ctx, c))
	buf.Reset()

	// add non-existing mount
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"foo", filepath.Join(u.Dir, "mount1")}))
	c = cli.NewContext(app, fs, nil)

	assert.Error(t, act.MountAdd(ctx, c))
	buf.Reset()

	// add some mounts
	assert.NoError(t, u.InitStore("mount1"))
	assert.NoError(t, u.InitStore("mount2"))
	assert.NoError(t, act.Store.AddMount(ctx, "mount1", u.StoreDir("mount1")))
	assert.NoError(t, act.Store.AddMount(ctx, "mount2", u.StoreDir("mount2")))

	// print mounts
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	c = cli.NewContext(app, fs, nil)

	assert.NoError(t, act.MountsPrint(ctx, c))
	buf.Reset()
}
