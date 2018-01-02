package action

import (
	"bytes"
	"context"
	"flag"
	"io/ioutil"
	"os"
	"testing"

	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestShow(t *testing.T) {
	td, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	act, err := newMock(ctx, td)
	assert.NoError(t, err)

	app := cli.NewApp()
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	if err := fs.Parse([]string{"foo"}); err != nil {
		t.Fatalf("Error: %s", err)
	}
	c := cli.NewContext(app, fs, nil)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	out := capture(t, func() error {
		return act.Show(ctx, c)
	})
	want := "0xDEADBEEF"
	assert.Equal(t, out, want)
}
