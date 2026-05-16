package main

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/fatih/color"
	"github.com/gopasspw/clipboard"
	"github.com/gopasspw/gopass/internal/action"
	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/backend/crypto/gpg"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/set"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestVersionPrinter(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	vp := makeVersionPrinter(buf, semver.Version{Major: 1})
	vp(nil)

	commit, _, _ := parseBuildInfo()

	assert.Contains(t, buf.String(), "gopass 1.0.0")
	assert.Contains(t, buf.String(), commit)
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

	ctx := config.NewContextInMemory()
	_, app := setupApp(ctx, semver.Version{})
	assert.NotNil(t, app)
}

// commandsWithError is a list of commands that return an error when
// invoked without arguments.
var commandsWithError = set.Map([]string{
	".age.identities.add",
	".age.identities.remove",
	".age.lock",
	".alias.add",
	".alias.remove",
	".alias.delete",
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
	".recipients.add",
	".recipients.remove",
	".show",
	".sum",
	".templates.edit",
	".templates.remove",
	".templates.show",
	".unclip",
	".reorg",
	".audit",
})

func TestGetCommands(t *testing.T) {
	u := gptest.NewUnitTester(t)

	buf := &bytes.Buffer{}
	color.NoColor = true

	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	cfg := config.NewInMemory()
	require.NoError(t, cfg.SetPath(u.StoreDir("")))

	clipboard.ForceUnsupported = true

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithInteractive(ctx, false)
	ctx = ctxutil.WithTerminal(ctx, false)
	ctx = ctxutil.WithHidden(ctx, true)
	ctx, err := backend.WithCryptoBackendString(ctx, "plain")
	require.NoError(t, err)
	ctx = ctxutil.WithAgePassphrase(ctx, "foobar")

	act, err := action.New(cfg, semver.Version{})
	require.NoError(t, err)

	app := &cli.Command{
		ExitErrHandler: func(_ context.Context, _ *cli.Command, _ error) {
			// suppress os.Exit during testing
		},
	}

	commands := getCommands(act, app)
	assert.Len(t, commands, 43)

	prefix := ""
	testCommands(t, ctx, app, commands, prefix)
}

func testCommands(t *testing.T, ctx context.Context, app *cli.Command, commands []*cli.Command, prefix string) {
	t.Helper()

	for _, cmd := range commands {
		if cmd.Name == "update" || cmd.Name == "agent" || cmd.Name == "doctor" {
			continue
		}

		if len(cmd.Commands) > 0 {
			testCommands(t, ctx, app, cmd.Commands, prefix+"."+cmd.Name)
		}

		if cmd.Before != nil {
			if _, err := runCmdBefore(ctx, cmd); err != nil {
				continue
			}
		}

		if cmd.ShellComplete != nil {
			cmd.ShellComplete(ctx, cmd)
		}

		if cmd.Action != nil {
			fullName := prefix + "." + cmd.Name
			if _, found := commandsWithError[fullName]; found {
				require.Error(t, runCmdAction(ctx, cmd), "Command %s should fail", fullName)

				continue
			}

			require.NoError(t, runCmdAction(ctx, cmd), "Command %s should not fail", fullName)
		}
	}
}

// runCmdAction invokes cmd.Action by running it through a parent cli.Command
// so that parsedArgs and flags are properly initialized.
func runCmdAction(ctx context.Context, cmd *cli.Command) error {
	wrapper := &cli.Command{
		ExitErrHandler: func(_ context.Context, _ *cli.Command, _ error) {
			// suppress os.Exit during testing
		},
		Commands: []*cli.Command{cmd},
	}

	return wrapper.Run(ctx, []string{"test", cmd.Name}) //nolint:wrapcheck
}

// runCmdBefore invokes cmd.Before by running it through a parent cli.Command.
func runCmdBefore(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	var capturedCtx context.Context
	var capturedErr error

	origBefore := cmd.Before
	cmd.Before = func(c context.Context, cmd *cli.Command) (context.Context, error) {
		capturedCtx, capturedErr = origBefore(c, cmd)

		return capturedCtx, capturedErr
	}
	// Use an action that does nothing so we can isolate the Before call
	origAction := cmd.Action
	cmd.Action = func(_ context.Context, _ *cli.Command) error { return nil }

	wrapper := &cli.Command{
		ExitErrHandler: func(_ context.Context, _ *cli.Command, _ error) {},
		Commands:       []*cli.Command{cmd},
	}
	_ = wrapper.Run(ctx, []string{"test", cmd.Name})

	// Restore original handlers
	cmd.Before = origBefore
	cmd.Action = origAction

	if capturedCtx == nil {
		capturedCtx = ctx
	}

	return capturedCtx, capturedErr
}

func TestInitContext(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	cfg := config.NewInMemory()

	ctx = initContext(ctx, cfg)
	assert.True(t, gpg.IsAlwaysTrust(ctx))
}
