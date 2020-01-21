package main

import (
	"context"
	"os"
	"strings"

	ap "github.com/gopasspw/gopass/pkg/action"
	"github.com/gopasspw/gopass/pkg/config"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/pkg/store/sub"
	"github.com/gopasspw/gopass/pkg/termio"

	"github.com/blang/semver"
	"gopkg.in/urfave/cli.v1"
)

func setupApp(ctx context.Context, sv semver.Version) *cli.App {
	// try to read config (if it exists)
	cfg := config.Load()

	// set config values
	ctx = initContext(ctx, cfg)

	// initialize action handlers
	action, err := ap.New(ctx, cfg, sv)
	if err != nil {
		out.Error(ctx, "No gpg binary found: %s", err)
		os.Exit(ap.ExitGPG)
	}

	// set some action callbacks
	if !cfg.Root.AutoImport {
		ctx = sub.WithImportFunc(ctx, termio.AskForKeyImport)
	}
	if !cfg.Root.NoConfirm {
		ctx = sub.WithRecipientFunc(ctx, action.ConfirmRecipients)
	}
	ctx = sub.WithFsckFunc(ctx, termio.AskForConfirmation)

	app := cli.NewApp()

	app.Name = name
	app.Version = sv.String()
	app.Usage = "The standard unix password manager - rewritten in Go"
	app.EnableBashCompletion = true
	app.BashComplete = func(c *cli.Context) {
		cli.DefaultAppComplete(c)
		action.Complete(ctx, c)
	}

	app.Action = func(c *cli.Context) error {
		if err := action.Initialized(withGlobalFlags(ctx, c), c); err != nil {
			return err
		}

		if strings.HasSuffix(os.Args[0], "native_host") || strings.HasSuffix(os.Args[0], "native_host.exe") {
			return action.JSONAPI(withGlobalFlags(ctx, c), c)
		}

		if c.Args().Present() {
			return action.Show(withGlobalFlags(ctx, c), c)
		}
		return action.List(withGlobalFlags(ctx, c), c)
	}

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "yes",
			Usage: "Assume yes on all yes/no questions or use the default on all others",
		},
		cli.BoolFlag{
			Name:  "clip, c",
			Usage: "Copy the first line of the secret into the clipboard",
		},
	}

	app.Commands = getCommands(ctx, action, app)
	return app
}
