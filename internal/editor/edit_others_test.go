//go:build !windows

package editor

import (
	"context"
	"os"
	"os/exec"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestEditor(t *testing.T) {
	// necessary for setting up the env
	u := gptest.NewGUnitTester(t)
	assert.NotNil(t, u)

	ctx := config.NewContextInMemory()
	touch, err := exec.LookPath("touch")
	require.NoError(t, err, os.Getenv("PATH"))

	want := "foobar"
	out, err := Invoke(ctx, touch, []byte(want))
	require.NoError(t, err)
	if string(out) != want {
		t.Errorf("%q != %q", string(out), want)
	}
}

// runWithFlags executes a function inside a cli.Command action so that flags are parsed.
func runWithFlags(flags map[string]string, fn func(context.Context, *cli.Command)) {
	cliFlags := make([]cli.Flag, 0, len(flags))
	args := make([]string, 0, len(flags)+1)
	args = append(args, "test")
	for k, v := range flags {
		cliFlags = append(cliFlags, &cli.StringFlag{Name: k})
		args = append(args, "--"+k+"="+v)
	}
	cmd := &cli.Command{
		Flags: cliFlags,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			fn(ctx, cmd)

			return nil
		},
	}
	_ = cmd.Run(context.Background(), args)
}

func TestGetEditor(t *testing.T) {
	t.Setenv("EDITOR", "")
	td := t.TempDir()
	t.Setenv("GOPASS_HOMEDIR", td)

	t.Run("--editor=fooed", func(t *testing.T) {
		var got string
		runWithFlags(map[string]string{"editor": "fooed"}, func(ctx context.Context, cmd *cli.Command) {
			got = Path(ctx, cmd)
		})
		assert.Equal(t, "fooed", got)
	})

	t.Run("/usr/bin/editor", func(t *testing.T) {
		var got string
		runWithFlags(nil, func(ctx context.Context, cmd *cli.Command) {
			got = Path(ctx, cmd)
		})
		pathed, err := exec.LookPath("editor")
		if err == nil {
			assert.Equal(t, pathed, got)
		}
	})

	t.Run("EDITOR", func(t *testing.T) {
		t.Setenv("EDITOR", "fooenv")
		var got string
		runWithFlags(nil, func(ctx context.Context, cmd *cli.Command) {
			got = Path(ctx, cmd)
		})
		assert.Equal(t, "fooenv", got)
	})

	t.Run("vi", func(t *testing.T) {
		t.Setenv("PATH", "/tmp")
		var got string
		runWithFlags(nil, func(ctx context.Context, cmd *cli.Command) {
			got = Path(ctx, cmd)
		})
		assert.Equal(t, "vi", got)
	})
}
