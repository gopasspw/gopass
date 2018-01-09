package action

import (
	"bytes"
	"context"
	"flag"
	"os"
	"os/exec"
	"testing"

	"github.com/justwatchcom/gopass/tests/gptest"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestEdit(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)
	act, err := newMock(ctx, u)
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
	assert.NoError(t, fs.Parse([]string{"foo"}))
	c = cli.NewContext(app, fs, nil)

	assert.Error(t, act.Edit(ctx, c))
	buf.Reset()

	// edit bar/foo with template
	assert.NoError(t, act.Store.SetTemplate(ctx, "bar", []byte("foobar")))
	buf.Reset()

	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"bar/foo"}))
	c = cli.NewContext(app, fs, nil)

	assert.Error(t, act.Edit(ctx, c))
	buf.Reset()
}

func TestEditor(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	act, err := newMock(ctx, u)
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

func TestGetEditor(t *testing.T) {
	app := cli.NewApp()

	// --editor=fooed
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	sf := cli.StringFlag{
		Name:  "editor",
		Usage: "editor",
	}
	assert.NoError(t, sf.ApplyWithError(fs))
	assert.NoError(t, fs.Parse([]string{"--editor", "fooed"}))
	c := cli.NewContext(app, fs, nil)

	assert.Equal(t, "fooed", getEditor(c))

	// EDITOR
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	c = cli.NewContext(app, fs, nil)
	assert.NoError(t, os.Setenv("EDITOR", "fooenv"))
	assert.Equal(t, "fooenv", getEditor(c))
	assert.NoError(t, os.Unsetenv("EDITOR"))

	// editor
	pathed, err := exec.LookPath("editor")
	if err == nil {
		assert.Equal(t, pathed, getEditor(c))
	}

	// vi
	op := os.Getenv("PATH")
	assert.NoError(t, os.Setenv("PATH", "/tmp"))
	assert.Equal(t, "vi", getEditor(c))
	assert.NoError(t, os.Setenv("PATH", op))
}
