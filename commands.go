package main

import (
	"context"
	"sort"

	ap "github.com/gopasspw/gopass/pkg/action"
	"github.com/gopasspw/gopass/pkg/action/binary"
	"github.com/gopasspw/gopass/pkg/action/create"
	"github.com/gopasspw/gopass/pkg/action/pwgen"
	"github.com/gopasspw/gopass/pkg/action/xc"
	"github.com/urfave/cli/v2"
)

func getCommands(ctx context.Context, action *ap.Action, app *cli.App) []*cli.Command {
	cmds := []*cli.Command{
		{
			Name:  "completion",
			Usage: "Bash and ZSH completion",
			Description: "" +
				"Source the output of this command with bash or zsh to get auto completion",
			Subcommands: []*cli.Command{{
				Name:   "bash",
				Usage:  "Source for auto completion in bash",
				Action: action.CompletionBash,
			}, {
				Name:  "zsh",
				Usage: "Source for auto completion in zsh",
				Action: func(c *cli.Context) error {
					return action.CompletionZSH(c, app)
				},
			}, {
				Name:  "fish",
				Usage: "Source for auto completion in fish",
				Action: func(c *cli.Context) error {
					return action.CompletionFish(c, app)
				},
			}, {
				Name:  "openbsdksh",
				Usage: "Source for auto completion in OpenBSD's ksh",
				Action: func(c *cli.Context) error {
					return action.CompletionOpenBSDKsh(c, app)
				},
			}},
		},
		{
			Name:        "jsonapi",
			Usage:       "Run and configure gopass as jsonapi e.g. for browser plugins",
			Description: "Setup and run gopass as native messaging hosts, e.g. for browser plugins.",
			Hidden:      false,
			Subcommands: []*cli.Command{
				{
					Name:        "listen",
					Usage:       "Listen and respond to messages via stdin/stdout",
					Description: "Gopass is started in listen mode from browser plugins using a wrapper specified in native messaging host manifests",
					Action:      action.JSONAPI,
					Before:      action.Initialized,
				},
				{
					Name:        "configure",
					Usage:       "Setup gopass native messaging manifest for selected browser",
					Description: "To access gopass from browser plugins, a native app manifest must be installed at the correct location",
					Action:      action.SetupNativeMessaging,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "browser",
							Usage: "One of 'chrome' and 'firefox'",
						},
						&cli.StringFlag{
							Name:  "path",
							Usage: "Path to install 'gopass_wrapper.sh' to",
						},
						&cli.StringFlag{
							Name:  "manifest-path",
							Usage: "Path to install 'com.justwatch.gopass.json' to",
						},
						&cli.BoolFlag{
							Name:  "global",
							Usage: "Install for all users, requires superuser rights",
						},
						&cli.StringFlag{
							Name:  "libpath",
							Usage: "Library path for global installation on linux. Default is /usr/lib",
						},
						&cli.StringFlag{
							Name:  "gopass-path",
							Usage: "Path to gopass binary. Default is auto detected",
						},
						&cli.BoolFlag{
							Name:  "print",
							Usage: "Print installation summary before creating any files",
							Value: true,
						},
					},
				},
			},
		},
		{
			Name:    "otp",
			Usage:   "Generate time- or hmac-based tokens",
			Aliases: []string{"totp", "hotp"},
			Hidden:  true,
			Description: "" +
				"Tries to parse an OTP URL (otpauth://). URL can be TOTP or HOTP. " +
				"The URL can be provided on its own line or on a key value line with a key named 'totp'.",
			Before:       action.Initialized,
			Action:       action.OTP,
			BashComplete: action.Complete,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "clip",
					Aliases: []string{"c"},
					Usage:   "Copy the time-based token into the clipboard",
				},
				&cli.StringFlag{
					Name:    "qr",
					Aliases: []string{"q"},
					Usage:   "Write QR code to FILE",
				},
				&cli.BoolFlag{
					Name:    "password",
					Aliases: []string{"o"},
					Usage:   "Only display the token",
				},
			},
		},
	}
	cmds = append(cmds, action.GetCommands()...)
	cmds = append(cmds, xc.GetCommands()...)
	cmds = append(cmds, create.GetCommands(action, action.Store)...)
	cmds = append(cmds, binary.GetCommands(action, action.Store)...)
	cmds = append(cmds, pwgen.GetCommands()...)
	sort.Slice(cmds, func(i, j int) bool { return cmds[i].Name < cmds[j].Name })
	return cmds
}
