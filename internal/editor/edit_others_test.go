// +build !windows

package editor

import (
	"context"
	"flag"
	"os"
	"os/exec"
	"runtime"
	"testing"

	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestEditor(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	touch, err := exec.LookPath("touch")
	require.NoError(t, err)

	want := "foobar"
	out, err := Invoke(ctx, touch, []byte(want))
	require.NoError(t, err)
	if string(out) != want {
		t.Errorf("%q != %q", string(out), want)
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
	require.NoError(t, sf.Apply(fs))
	require.NoError(t, fs.Parse([]string{"--editor", "fooed"}))
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
	if runtime.GOOS == "windows" {
		assert.Equal(t, "notepad.exe", Path(c))

	} else {
		assert.Equal(t, "vi", Path(c))

	}
	assert.NoError(t, os.Setenv("PATH", op))
}
