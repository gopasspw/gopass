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

func TestMounts(t *testing.T) {
	td, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	act, err := newMock(ctx, td)
	assert.NoError(t, err)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
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
	if err := fs.Parse([]string{"foo"}); err != nil {
		t.Fatalf("Error: %s", err)
	}
	c = cli.NewContext(app, fs, nil)

	assert.NoError(t, act.MountRemove(ctx, c))
	buf.Reset()

	// add non-existing mount
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	if err := fs.Parse([]string{"foo", filepath.Join(td, "mount1")}); err != nil {
		t.Fatalf("Error: %s", err)
	}
	c = cli.NewContext(app, fs, nil)

	assert.Error(t, act.MountAdd(ctx, c))
	buf.Reset()

}
