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

	rl.Config.AutoComplete = s.prefixCompleter(c)

	for {
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
		if strings.ToLower(args[0]) == "quit" {
			break
		}
		if err := c.App.RunContext(c.Context, append([]string{"gopass"}, args...)); err != nil {
			continue
		}
	}

	return nil
}
