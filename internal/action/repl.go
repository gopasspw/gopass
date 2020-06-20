package action

import (
	"context"
	"io"
	"strings"

	"github.com/chzyer/readline"
	"github.com/gopasspw/gopass/internal/debug"
	"github.com/gopasspw/gopass/internal/out"
	shellquote "github.com/kballard/go-shellquote"
	"github.com/urfave/cli/v2"
)

func (s *Action) entriesForCompleter(ctx context.Context) ([]readline.PrefixCompleterInterface, error) {
	args := []readline.PrefixCompleterInterface{}
	list, err := s.Store.List(ctx, 0)
	if err != nil {
		return args, err
	}
	for _, v := range list {
		args = append(args, readline.PcItem(v))
	}
	return args, nil
}

func (s *Action) replCompleteRecipients(ctx context.Context, cmd *cli.Command) []readline.PrefixCompleterInterface {
	subCmds := []readline.PrefixCompleterInterface{}
	if cmd.Name == "remove" {
		for _, r := range s.recipientsList(ctx) {
			subCmds = append(subCmds, readline.PcItem(r))
		}
	}
	args := []readline.PrefixCompleterInterface{}
	args = append(args, readline.PcItem(cmd.Name, subCmds...))
	for _, alias := range cmd.Aliases {
		args = append(args, readline.PcItem(alias, subCmds...))
	}
	return args
}

func (s *Action) replCompleteTemplates(ctx context.Context, cmd *cli.Command) []readline.PrefixCompleterInterface {
	subCmds := []readline.PrefixCompleterInterface{}
	for _, r := range s.templatesList(ctx) {
		subCmds = append(subCmds, readline.PcItem(r))
	}
	args := []readline.PrefixCompleterInterface{}
	args = append(args, readline.PcItem(cmd.Name, subCmds...))
	for _, alias := range cmd.Aliases {
		args = append(args, readline.PcItem(alias, subCmds...))
	}
	return args
}

func (s *Action) prefixCompleter(c *cli.Context) *readline.PrefixCompleter {
	secrets, err := s.entriesForCompleter(c.Context)
	if err != nil {
		debug.Log("failed to list secrets: %s", err)
	}
	cmds := []readline.PrefixCompleterInterface{}
	for _, cmd := range c.App.Commands {
		if cmd.Hidden {
			continue
		}
		subCmds := []readline.PrefixCompleterInterface{}
		switch cmd.Name {
		case "config":
			for _, k := range s.configKeys() {
				subCmds = append(subCmds, readline.PcItem(k))
			}
		case "recipients":
			subCmds = append(subCmds, s.replCompleteRecipients(c.Context, cmd)...)
		case "templates":
			subCmds = append(subCmds, s.replCompleteTemplates(c.Context, cmd)...)
		case "cat":
			fallthrough
		case "delete":
			fallthrough
		case "edit":
			fallthrough
		case "generate":
			fallthrough
		case "history":
			fallthrough
		case "list":
			fallthrough
		case "move":
			fallthrough
		case "otp":
			fallthrough
		case "show":
			subCmds = append(subCmds, secrets...)
		default:
		}
		for _, scmd := range cmd.Subcommands {
			subCmds = append(subCmds, readline.PcItem(scmd.Name))
		}
		cmds = append(cmds, readline.PcItem(cmd.Name, subCmds...))
		for _, alias := range cmd.Aliases {
			cmds = append(cmds, readline.PcItem(alias, subCmds...))
		}
	}
	return readline.NewPrefixCompleter(cmds...)
}

// REPL implements a read-execute-print-line shell
// with readline support and autocompletion.
func (s *Action) REPL(c *cli.Context) error {
	c.App.ExitErrHandler = func(c *cli.Context, err error) {
		if err == nil {
			return
		}
		out.Red(c.Context, "Error: %s", err)
	}
	rl, err := readline.New("gopass> ")
	if err != nil {
		return err
	}
	defer rl.Close()

READ:
	for {
		rl.Config.AutoComplete = s.prefixCompleter(c)
		line, err := rl.Readline()
		if err != nil {
			if err == io.EOF {
				break
			}
			debug.Log("Readline error: %s", err)
		}
		args, err := shellquote.Split(line)
		if err != nil {
			out.Red(c.Context, "Error: %s", err)
			continue
		}
		if len(args) < 1 {
			continue
		}
		switch strings.ToLower(args[0]) {
		case "quit":
			break READ
		case "clear":
			readline.ClearScreen(stdout)
			continue
		default:
		}
		// need to reinitialize the config to pick up any changes from the
		// previous iteration
		// TODO: this means the context will grow with every loop. Eventually
		// this might lead to memory issues so we should see if we can optimize it.
		c.Context = s.cfg.WithContext(c.Context)
		if err := c.App.RunContext(c.Context, append([]string{"gopass"}, args...)); err != nil {
			continue
		}
	}
	return nil
}
