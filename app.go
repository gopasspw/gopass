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
	"github.com/urfave/cli/v2"
)

func setupApp(ctx context.Context, sv semver.Version) (context.Context, *cli.App) {
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
		action.Complete(c)
	}

	app.Action = func(c *cli.Context) error {
		if err := action.Initialized(c); err != nil {
			return err
		}

		if strings.HasSuffix(os.Args[0], "native_host") || strings.HasSuffix(os.Args[0], "native_host.exe") {
			return action.JSONAPI(c)
		}

		if c.Args().Present() {
			return action.Show(c)
		}
		return action.List(c)
	}

	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:  "yes",
			Usage: "Assume yes on all yes/no questions or use the default on all others",
		},
		&cli.BoolFlag{
			Name:  "clip, c",
			Usage: "Copy the first line of the secret into the clipboard",
		},
		&cli.BoolFlag{
			Name:  "alsoclip, C",
			Usage: "Copy the first line of the secret into the clipboard and show everything",
		},
	}

	app.Commands = getCommands(ctx, action, app)
	return ctx, app
}
