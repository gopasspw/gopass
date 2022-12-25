//go:build !windows
// +build !windows

package editor

import (
	"context"
	"flag"
	"os"
	"os/exec"
	"testing"

	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestEditor(t *testing.T) {
	// necessary for setting up the env
	u := gptest.NewGUnitTester(t)
	assert.NotNil(t, u)

	ctx := context.Background()
	touch, err := exec.LookPath("touch")
	require.NoError(t, err, os.Getenv("PATH"))

	want := "foobar"
	out, err := Invoke(ctx, touch, []byte(want))
	require.NoError(t, err)
	if string(out) != want {
		t.Errorf("%q != %q", string(out), want)
	}
}

func TestGetEditor(t *testing.T) {
	app := cli.NewApp()

	t.Setenv("EDITOR", "")

	t.Run("--editor=fooed", func(t *testing.T) {
		fs := flag.NewFlagSet("default", flag.ContinueOnError)
		sf := cli.StringFlag{
			Name:  "editor",
			Usage: "editor",
		}
		require.NoError(t, sf.Apply(fs))
		require.NoError(t, fs.Parse([]string{"--editor", "fooed"}))
		c := cli.NewContext(app, fs, nil)

		assert.Equal(t, "fooed", Path(c))
	})

	t.Run("/usr/bin/editor", func(t *testing.T) {
		fs := flag.NewFlagSet("default", flag.ContinueOnError)
		c := cli.NewContext(app, fs, nil)
		pathed, err := exec.LookPath("editor")
		if err == nil {
			assert.Equal(t, pathed, Path(c))
		}
	})

	t.Run("EDITOR", func(t *testing.T) {
		fs := flag.NewFlagSet("default", flag.ContinueOnError)
		c := cli.NewContext(app, fs, nil)
		t.Setenv("EDITOR", "fooenv")
		assert.Equal(t, "fooenv", Path(c))
	})

	t.Run("vi", func(t *testing.T) {
		fs := flag.NewFlagSet("default", flag.ContinueOnError)
		c := cli.NewContext(app, fs, nil)
		t.Setenv("PATH", "/tmp")
		assert.Equal(t, "vi", Path(c))
	})
}
