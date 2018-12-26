package action

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/cui"
	"github.com/gopasspw/gopass/pkg/out"

	"github.com/schollz/closestmatch"
	"github.com/urfave/cli"
)

// Find a string in the secret file's name
func (s *Action) Find(ctx context.Context, c *cli.Context) error {
	if c.IsSet("clip") {
		ctx = WithClip(ctx, c.Bool("clip"))
	}

	if !c.Args().Present() {
		return ExitError(ctx, ExitUsage, nil, "Usage: %s find <NEEDLE>", s.Name)
	}

	return s.find(ctx, c, c.Args().First(), s.show)
}

// see action.show - context, cli context, name, key, rescurse
type showFunc func(context.Context, *cli.Context, string, string, bool) error

func (s *Action) find(ctx context.Context, c *cli.Context, needle string, cb showFunc) error {
	// get all existing entries
	haystack, err := s.Store.List(ctx, 0)
	if err != nil {
		return ExitError(ctx, ExitList, err, "failed to list store: %s", err)
	}

	// filter our the ones from the haystack matching the needle
	needle = strings.ToLower(needle)
	choices := filter(haystack, needle)

	// if we have an exact match print it
	if len(choices) == 1 {
		out.Green(ctx, "Found exact match in '%s'", choices[0])
		return cb(ctx, c, choices[0], "", false)
	}

	// if we don't have a match yet try a fuzzy search
	if len(choices) < 1 && ctxutil.IsFuzzySearch(ctx) {
		// try fuzzy match
		cm := closestmatch.New(haystack, []int{2})
		choices = cm.ClosestN(needle, 5)
	}

	// if there are still no results we abort
	if len(choices) < 1 {
		return ExitError(ctx, ExitNotFound, nil, "no results found")
	}

	// do not invoke wizard if not printing to terminal or if
	// gopass find/search was invoked directly (for scripts)
	if !ctxutil.IsTerminal(ctx) || (c != nil && c.Command.Name == "find") {
		for _, value := range choices {
			out.Print(ctx, value)
		}
		return nil
	}

	return s.findSelection(ctx, c, choices, needle, cb)
}

// findSelection runs a wizard that lets the user select an entry
func (s *Action) findSelection(ctx context.Context, c *cli.Context, choices []string, needle string, cb showFunc) error {
	sort.Strings(choices)
	act, sel := cui.GetSelection(ctx, "Found secrets - Please select an entry", "<↑/↓> to change the selection, <→> to show, <←> to copy, <s> to sync, <e> to edit, <ESC> to quit", choices)
	out.Debug(ctx, "Action: %s - Selection: %d", act, sel)
	switch act {
	case "default":
		// display or copy selected entry
		fmt.Fprintln(stdout, choices[sel])
		return cb(ctx, c, choices[sel], "", false)
	case "copy":
		// display selected entry
		fmt.Fprintln(stdout, choices[sel])
		return cb(WithClip(ctx, true), c, choices[sel], "", false)
	case "show":
		// display selected entry
		fmt.Fprintln(stdout, choices[sel])
		return cb(WithClip(ctx, false), c, choices[sel], "", false)
	case "sync":
		// run sync and re-run show/find workflow
		if err := s.Sync(ctx, c); err != nil {
			return err
		}
		return cb(ctx, c, needle, "", true)
	case "edit":
		// edit selected entry
		fmt.Fprintln(stdout, choices[sel])
		return s.edit(ctx, c, choices[sel])
	default:
		return ExitError(ctx, ExitAborted, nil, "user aborted")
	}
}

func filter(l []string, needle string) []string {
	choices := make([]string, 0, 10)
	for _, value := range l {
		if strings.Contains(strings.ToLower(value), needle) {
			choices = append(choices, value)
		}
	}
	return choices
}
