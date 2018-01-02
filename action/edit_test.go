package action

import (
	"bytes"
	"context"
	"flag"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestEdit(t *testing.T) {
	td, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)
	act, err := newMock(ctx, td)
	assert.NoError(t, err)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	app := cli.NewApp()

	// edit
	c := cli.NewContext(app, flag.NewFlagSet("default", flag.ContinueOnError), nil)
	if err := act.Edit(ctx, c); err == nil || err.Error() != "Usage: action.test edit secret" {
		t.Errorf("Should fail")
	}

	// edit foo
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	if err := fs.Parse([]string{"foo"}); err != nil {
		t.Fatalf("Error: %s", err)
	}
	c = cli.NewContext(app, fs, nil)

	assert.Error(t, act.Edit(ctx, c))
	buf.Reset()
}

func TestEditor(t *testing.T) {
	td, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()

	ctx := context.Background()
	act, err := newMock(ctx, td)
	assert.NoError(t, err)

	touch, err := exec.LookPath("touch")
	assert.NoError(t, err)

	want := "foobar"
	out, err := act.editor(ctx, touch, []byte(want))
	assert.NoError(t, err)
	if string(out) != want {
		t.Errorf("'%s' != '%s'", string(out), want)
	}
}
