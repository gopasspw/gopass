package main

import (
	"bytes"
	"context"
	"flag"
	"os"
	"testing"

	"github.com/atotto/clipboard"
	"github.com/blang/semver"
	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/action"
	"github.com/justwatchcom/gopass/backend"
	"github.com/justwatchcom/gopass/config"
	"github.com/justwatchcom/gopass/tests/gptest"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

var commandsWithError = map[string]struct{}{
	".audit":                 {},
	".audit.hibp":            {},
	".binary.cat":            {},
	".binary.copy":           {},
	".binary.move":           {},
	".binary.sum":            {},
	".clone":                 {},
	".copy":                  {},
	".create":                {},
	".delete":                {},
	".edit":                  {},
	".find":                  {},
	".generate":              {},
	".grep":                  {},
	".init":                  {},
	".insert":                {},
	".mounts.add":            {},
	".mounts.remove":         {},
	".move":                  {},
	".otp":                   {},
	".recipients.add":        {},
	".recipients.remove":     {},
	".setup":                 {},
	".show":                  {},
	".templates.edit":        {},
	".templates.remove":      {},
	".templates.show":        {},
	".xc.export":             {},
	".xc.export-private-key": {},
	".xc.generate":           {},
	".xc.import":             {},
	".xc.import-private-key": {},
	".xc.remove":             {},
	".unclip":                {},
}

func TestGetCommands(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	buf := &bytes.Buffer{}
	out.Stdout = buf
	color.NoColor = true
	defer func() {
		out.Stdout = os.Stdout
	}()

	cfg := config.New()
	cfg.Root.Path = u.StoreDir("")

	clipboard.Unsupported = true

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithInteractive(ctx, false)
	ctx = ctxutil.WithTerminal(ctx, false)
	ctx = out.WithHidden(ctx, true)
	ctx = backend.WithSyncBackendString(ctx, "gitmock")
	ctx = backend.WithCryptoBackendString(ctx, "gpgmock")

	act, err := action.New(ctx, cfg, semver.Version{})
	assert.NoError(t, err)

	app := cli.NewApp()
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	c := cli.NewContext(app, fs, nil)

	commands := getCommands(ctx, act, app)
	assert.Equal(t, 30, len(commands))

	prefix := ""
	testCommands(t, c, commands, prefix)
}

func testCommands(t *testing.T, c *cli.Context, commands []cli.Command, prefix string) {
	for _, cmd := range commands {
		if cmd.Name == "agent" || cmd.Name == "update" {
			// the agent command is blocking
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
		if cmd.Action != nil {
			fullName := prefix + "." + cmd.Name
			if av, ok := cmd.Action.(func(c *cli.Context) error); ok {
				if _, found := commandsWithError[fullName]; found {
					assert.Error(t, av(c), fullName)
					continue
				}
				assert.NoError(t, av(c), fullName)
			}
		}
	}
}
