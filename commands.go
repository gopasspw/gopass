package main

import (
	"context"
	"sort"

	ap "github.com/gopasspw/gopass/internal/action"
	"github.com/gopasspw/gopass/internal/action/binary"
	"github.com/gopasspw/gopass/internal/action/create"
	"github.com/gopasspw/gopass/internal/action/pwgen"
	"github.com/gopasspw/gopass/internal/action/xc"
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
	}
	cmds = append(cmds, action.GetCommands()...)
	cmds = append(cmds, xc.GetCommands()...)
	cmds = append(cmds, create.GetCommands(action, action.Store)...)
	cmds = append(cmds, binary.GetCommands(action, action.Store)...)
	cmds = append(cmds, pwgen.GetCommands()...)
	sort.Slice(cmds, func(i, j int) bool { return cmds[i].Name < cmds[j].Name })
	return cmds
}
