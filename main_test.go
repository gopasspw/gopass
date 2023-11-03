package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"
	"testing"

	"github.com/atotto/clipboard"
	"github.com/blang/semver/v4"
	"github.com/fatih/color"
	"github.com/gopasspw/gopass/internal/action"
	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/backend/crypto/gpg"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/set"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestVersionPrinter(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	vp := makeVersionPrinter(buf, semver.Version{Major: 1})
	vp(nil)
	assert.Equal(t, fmt.Sprintf("gopass 1.0.0 %s %s %s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH), buf.String())
}

func TestGetVersion(t *testing.T) {
	t.Parallel()

	version = "1.9.0"

	if getVersion().LT(semver.Version{Major: 1, Minor: 9}) {
		t.Errorf("invalid version")
	}
}

func TestSetupApp(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	_, app := setupApp(ctx, semver.Version{})
	assert.NotNil(t, app)
}

// commandsWithError is a list of commands that return an error when
// invoked without arguments.
var commandsWithError = set.Map([]string{
	".age.identities.add",
	".age.identities.remove",
	".alias.add",
	".alias.remove",
	".alias.delete",
	".audit",
	".cat",
	".clone",
	".copy",
	".create",
	".delete",
	".edit",
	".env",
	".find",
	".fscopy",
	".fsmove",
	".generate",
	".git",
	".git.push",
	".git.pull",
	".git.status",
	".git.remote.add",
	".git.remote.remove",
	".grep",
	".history",
	".init",
	".insert",
	".link",
	".merge",
	".mounts.add",
	".mounts.remove",
	".move",
	".otp",
	".process",
	".rcs.status",
	".recipients.add",
	".recipients.remove",
	".show",
	".sum",
	".templates.edit",
	".templates.remove",
	".templates.show",
	".unclip",
})

func TestGetCommands(t *testing.T) {
	u := gptest.NewUnitTester(t)

	buf := &bytes.Buffer{}
	color.NoColor = true

	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	cfg := config.NewNoWrites()
	require.NoError(t, cfg.SetPath(u.StoreDir("")))

	clipboard.Unsupported = true

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithInteractive(ctx, false)
	ctx = ctxutil.WithTerminal(ctx, false)
	ctx = ctxutil.WithHidden(ctx, true)
	ctx = backend.WithCryptoBackendString(ctx, "plain")

	act, err := action.New(cfg, semver.Version{})
	require.NoError(t, err)

	app := cli.NewApp()
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	c := cli.NewContext(app, fs, nil)
	c.Context = ctx

	commands := getCommands(act, app)
	assert.Len(t, commands, 41)

	prefix := ""
	testCommands(t, c, commands, prefix)
}

func testCommands(t *testing.T, c *cli.Context, commands []*cli.Command, prefix string) {
	t.Helper()

	for _, cmd := range commands {
		if cmd.Name == "update" {
			continue
		}

		if len(cmd.Subcommands) > 0 {
			testCommands(t, c, cmd.Subcommands, prefix+"."+cmd.Name)
		}

		if cmd.Before != nil {
			if err := cmd.Before(c); err != nil {
				continue
			}
		}

		if cmd.BashComplete != nil {
			cmd.BashComplete(c)
		}

		if cmd.Action != nil {
			fullName := prefix + "." + cmd.Name
			if _, found := commandsWithError[fullName]; found {
				require.Error(t, cmd.Action(c), fullName)

				continue
			}

			require.NoError(t, cmd.Action(c), fullName)
		}
	}
}

func TestInitContext(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg := config.NewNoWrites()

	ctx = initContext(ctx, cfg)
	assert.True(t, gpg.IsAlwaysTrust(ctx))
}
