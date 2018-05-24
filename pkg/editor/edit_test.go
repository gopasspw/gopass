package editor

import (
	"bytes"
	"context"
	"flag"
	"os"
	"os/exec"
	"testing"

	"github.com/justwatchcom/gopass/pkg/ctxutil"
	"github.com/justwatchcom/gopass/pkg/out"
	"github.com/justwatchcom/gopass/tests/gptest"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestEdit(t *testing.T) {
	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	_, err := Invoke(ctx, "true", []byte{})
	assert.Error(t, err)
	buf.Reset()
}

func TestEditor(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	touch, err := exec.LookPath("touch")
	assert.NoError(t, err)

	want := "foobar"
	out, err := Invoke(ctx, touch, []byte(want))
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

	assert.Equal(t, "fooed", Path(c))

	// EDITOR
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	c = cli.NewContext(app, fs, nil)
	assert.NoError(t, os.Setenv("EDITOR", "fooenv"))
	assert.Equal(t, "fooenv", Path(c))
	assert.NoError(t, os.Unsetenv("EDITOR"))

	// editor
	pathed, err := exec.LookPath("editor")
	if err == nil {
		assert.Equal(t, pathed, Path(c))
	}

	// vi
	op := os.Getenv("PATH")
	assert.NoError(t, os.Setenv("PATH", "/tmp"))
	assert.Equal(t, "vi", Path(c))
	assert.NoError(t, os.Setenv("PATH", op))
}
