package action

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/ergochat/readline"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/tree"
	"github.com/gopasspw/gopass/pkg/debug"
	shellquote "github.com/kballard/go-shellquote"
	"github.com/urfave/cli/v2"
)

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
		case <-c.Done():
			return fmt.Errorf("user aborted")
		default:
		}

		// we need to update the completer on every loop since
		// the list of secrets may have changed, e.g. due to
		// the user adding a new secret.
		cfg := rl.GetConfig()
		cfg.AutoComplete = s.newGopassCompleter(c)
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

// escapeEntry escapes special shell characters in a secret name so that
// tab-completed values are safe to use on the REPL command line.
// Spaces, quotes, backslashes and other special chars are backslash-escaped.
func escapeEntry(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch r {
		case ' ', '\\', '\'', '"', '(', ')', '<', '>', '&', ';', '#', '|', '*', '?':
			b.WriteByte('\\')
		}
		b.WriteRune(r)
	}

	return b.String()
}

// unescapeEntry reverses escapeEntry — it removes backslash escapes so
// the raw entry name can be matched against store contents.
func unescapeEntry(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	escaped := false
	for _, r := range s {
		if escaped {
			b.WriteRune(r)
			escaped = false

			continue
		}
		if r == '\\' {
			escaped = true

			continue
		}
		b.WriteRune(r)
	}

	return b.String()
}

// completionSpec describes what candidates a command should complete against.
type completionSpec int

const (
	completeNone       completionSpec = iota
	completeEntries                   // secret entries
	completeConfig                    // config keys
	completeRecipients                // recipients (for "recipients remove")
	completeTemplates                 // templates
	completeSubCmds                   // subcommands only (no entries)
)

// gopassCompleter implements readline.AutoCompleter with proper support
// for secret names containing spaces.
type gopassCompleter struct {
	cmdSpecs   map[string]completionSpec
	subCmds    map[string][]string
	entries    []string
	configKeys []string
	recipients []string
	templates  []string
	commands   []string
}

// Do implements readline.AutoCompleter.
func (g *gopassCompleter) Do(line []rune, pos int) ([][]rune, int) {
	text := string(line[:pos])

	cmd, argPart, hasCmd := g.parseLine(text)

	if !hasCmd {
		return g.completeFromList(g.commands, cmd, false)
	}

	spec, ok := g.cmdSpecs[cmd]
	if !ok {
		spec = completeNone
	}

	var candidates []string

	switch spec {
	case completeEntries:
		candidates = g.entries
	case completeConfig:
		candidates = g.configKeys
	case completeRecipients:
		candidates = g.recipients
	case completeTemplates:
		candidates = g.templates
	case completeSubCmds:
		// nothing extra
	case completeNone:
		// nothing
	}

	if subs, ok := g.subCmds[cmd]; ok {
		candidates = append(candidates, subs...)
	}

	if len(candidates) == 0 {
		return nil, 0
	}

	needsEscape := spec == completeEntries || spec == completeTemplates

	return g.completeFromList(candidates, argPart, needsEscape)
}

// parseLine splits the REPL input into the command token and the argument
// portion. Flag tokens (starting with '-') are skipped.
func (g *gopassCompleter) parseLine(text string) (string, string, bool) {
	i := 0
	n := len(text)

	for i < n && text[i] == ' ' {
		i++
	}

	cmdStart := -1
	cmdEnd := -1
	for i < n {
		for i < n && text[i] == ' ' {
			i++
		}
		if i >= n {
			break
		}
		tokStart := i
		for i < n && text[i] != ' ' {
			i++
		}
		tok := text[tokStart:i]
		if strings.HasPrefix(tok, "-") {
			continue
		}
		cmdStart = tokStart
		cmdEnd = i

		break
	}

	if cmdStart == -1 {
		return strings.TrimSpace(text), "", false
	}

	cmd := text[cmdStart:cmdEnd]

	if cmdEnd >= len(text) || text[cmdEnd] != ' ' {
		return cmd, "", false
	}

	argPart := g.stripFlags(text[cmdEnd:])

	return cmd, argPart, true
}

// stripFlags removes flag tokens (words starting with '-') from the input,
// preserving spaces and backslash escapes.
func (g *gopassCompleter) stripFlags(s string) string {
	var b strings.Builder
	i := 0
	n := len(s)

	leadingSpace := false
	for i < n && s[i] == ' ' {
		leadingSpace = true
		i++
	}

	if !leadingSpace {
		return s
	}

	result := ""
	for i < n {
		spaceStart := i
		for i < n && s[i] == ' ' {
			i++
		}
		if i >= n {
			result += s[spaceStart:]

			break
		}

		tokStart := i
		escaped := false
		for i < n {
			if escaped {
				escaped = false
				i++

				continue
			}
			if s[i] == '\\' {
				escaped = true
				i++

				continue
			}
			if s[i] == ' ' {
				break
			}
			i++
		}

		tok := s[tokStart:i]
		if strings.HasPrefix(tok, "-") {
			continue
		}

		result += s[spaceStart:i]
	}

	if result == "" {
		if n > 0 && s[n-1] == ' ' {
			return " "
		}

		return ""
	}

	b.WriteString(result)

	return b.String()
}

// completeFromList finds all candidates matching the given prefix and returns
// readline-compatible completion results.
func (g *gopassCompleter) completeFromList(candidates []string, prefix string, needsEscape bool) ([][]rune, int) {
	trimmedPrefix := strings.TrimLeft(prefix, " ")
	rawPrefix := unescapeEntry(trimmedPrefix)

	var matches [][]rune
	for _, c := range candidates {
		if !strings.HasPrefix(strings.ToLower(c), strings.ToLower(rawPrefix)) {
			continue
		}

		var displayCandidate, escapedPrefix string
		if needsEscape {
			displayCandidate = escapeEntry(c)
			escapedPrefix = escapeEntry(rawPrefix)
		} else {
			displayCandidate = c
			escapedPrefix = rawPrefix
		}

		if !strings.HasPrefix(displayCandidate, escapedPrefix) {
			continue
		}

		suffix := displayCandidate[len(escapedPrefix):] + " "
		matches = append(matches, []rune(suffix))
	}

	if len(matches) == 0 {
		return nil, 0
	}

	sort.Slice(matches, func(i, j int) bool {
		return string(matches[i]) < string(matches[j])
	})

	return matches, len([]rune(trimmedPrefix))
}

// newGopassCompleter builds a gopassCompleter from the current app state.
func (s *Action) newGopassCompleter(c *cli.Context) *gopassCompleter {
	entries, err := s.Store.List(c.Context, tree.INF)
	if err != nil {
		debug.Log("failed to list secrets: %s", err)
		entries = nil
	}

	gc := &gopassCompleter{
		cmdSpecs: make(map[string]completionSpec),
		subCmds:  make(map[string][]string),
		entries:  entries,
	}

	entryCommands := map[string]bool{
		"cat": true, "delete": true, "edit": true, "generate": true,
		"history": true, "list": true, "move": true, "otp": true,
		"show": true,
	}

	for _, cmd := range c.App.Commands {
		if cmd.Hidden {
			continue
		}

		var spec completionSpec
		switch {
		case entryCommands[cmd.Name]:
			spec = completeEntries
		case cmd.Name == "config":
			spec = completeConfig
			if gc.configKeys == nil {
				gc.configKeys = s.configKeys()
			}
		case cmd.Name == "recipients":
			spec = completeRecipients
			if gc.recipients == nil {
				gc.recipients = s.recipientsList(c.Context)
			}
		case cmd.Name == "templates":
			spec = completeTemplates
			if gc.templates == nil {
				gc.templates = s.templatesList(c.Context)
			}
		default:
			spec = completeSubCmds
		}

		gc.cmdSpecs[cmd.Name] = spec
		gc.commands = append(gc.commands, cmd.Name)
		for _, alias := range cmd.Aliases {
			gc.cmdSpecs[alias] = spec
			gc.commands = append(gc.commands, alias)
		}

		if len(cmd.Subcommands) > 0 {
			var subs []string
			for _, scmd := range cmd.Subcommands {
				subs = append(subs, scmd.Name)
			}
			gc.subCmds[cmd.Name] = subs
			for _, alias := range cmd.Aliases {
				gc.subCmds[alias] = subs
			}
		}
	}

	sort.Strings(gc.commands)

	return gc
}
