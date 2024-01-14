package action

import (
	"context"
	"fmt"
	"strings"

	"github.com/ergochat/readline"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/tree"
	"github.com/gopasspw/gopass/pkg/debug"
	shellquote "github.com/kballard/go-shellquote"
	"github.com/urfave/cli/v2"
)

func (s *Action) entriesForCompleter(ctx context.Context) ([]*readline.PrefixCompleter, error) {
	args := []*readline.PrefixCompleter{}
	list, err := s.Store.List(ctx, tree.INF)
	if err != nil {
		return args, err
	}
	for _, v := range list {
		args = append(args, readline.PcItem(v))
	}

	return args, nil
}

func (s *Action) replCompleteRecipients(ctx context.Context, cmd *cli.Command) []*readline.PrefixCompleter {
	subCmds := []*readline.PrefixCompleter{}
	if cmd.Name == "remove" {
		for _, r := range s.recipientsList(ctx) {
			subCmds = append(subCmds, readline.PcItem(r))
		}
	}
	args := []*readline.PrefixCompleter{}
	args = append(args, readline.PcItem(cmd.Name, subCmds...))
	for _, alias := range cmd.Aliases {
		args = append(args, readline.PcItem(alias, subCmds...))
	}

	return args
}

func (s *Action) replCompleteTemplates(ctx context.Context, cmd *cli.Command) []*readline.PrefixCompleter {
	subCmds := []*readline.PrefixCompleter{}
	for _, r := range s.templatesList(ctx) {
		subCmds = append(subCmds, readline.PcItem(r))
	}
	args := []*readline.PrefixCompleter{}
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
	cmds := []*readline.PrefixCompleter{}
	for _, cmd := range c.App.Commands {
		if cmd.Hidden {
			continue
		}
		subCmds := []*readline.PrefixCompleter{}
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
		out.Errorf(c.Context, "%s", err)
	}

	out.Printf(c.Context, logo)
	out.Printf(c.Context, "🌟 Welcome to gopass!")
	out.Printf(c.Context, "⚠ This is the built-in shell. Type 'help' for a list of commands.")

	rl, err := readline.New("gopass> ")
	if err != nil {
		return err
	}

	defer func() {
		_ = rl.Close()
	}()

READ:
	for {
		// check for context cancellation
		select {
		case <-c.Context.Done():
			return fmt.Errorf("user aborted")
		default:
		}

		// we need to update the completer on every loop since
		// the list of secrets may have changed, e.g. due to
		// the user adding a new secret.
		cfg := rl.GetConfig()
		cfg.AutoComplete = s.prefixCompleter(c)
		if err := rl.SetConfig(cfg); err != nil {
			debug.Log("Failed to set readline config: %s", err)

			break
		}

		line, err := rl.Readline()
		if err != nil {
			debug.Log("Readline error: %s", err)

			break
		}
		args, err := shellquote.Split(line)
		if err != nil {
			out.Printf(c.Context, "Error: %s", err)

			continue
		}
		if len(args) < 1 {
			continue
		}
		switch strings.ToLower(args[0]) {
		case "quit":
			break READ
		case "lock":
			s.replLock(c.Context)

			continue
		case "clear":
			rl.ClearScreen()

			continue
		default:
		}

		if err := c.App.RunContext(c.Context, append([]string{"gopass"}, args...)); err != nil {
			continue
		}
	}

	return nil
}

func (s *Action) replLock(ctx context.Context) {
	if err := s.Store.Lock(); err != nil {
		out.Errorf(ctx, "Failed to lock stores: %s", err)

		return
	}
	out.OKf(ctx, "Locked")
}
