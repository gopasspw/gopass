package action

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/gopasspw/gopass/internal/action/exit"
	"github.com/gopasspw/gopass/internal/cui"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/tree"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/schollz/closestmatch"
	"github.com/urfave/cli/v2"
)

// Find runs find without fuzzy search.
func (s *Action) Find(c *cli.Context) error {
	return s.findCmd(c, nil, false)
}

// FindFuzzy runs find with fuzzy search.
func (s *Action) FindFuzzy(c *cli.Context) error {
	return s.findCmd(c, s.show, true)
}

func (s *Action) findCmd(c *cli.Context, cb showFunc, fuzzy bool) error {
	ctx := ctxutil.WithGlobalFlags(c)
	if c.IsSet("clip") {
		ctx = WithOnlyClip(ctx, c.Bool("clip"))
		ctx = WithClip(ctx, c.Bool("clip"))
	}

	if c.IsSet("unsafe") {
		ctx = ctxutil.WithForce(ctx, c.Bool("unsafe"))
	}

	if !c.Args().Present() {
		return exit.Error(exit.Usage, nil, "Usage: %s find <pattern>", s.Name)
	}

	return s.find(ctx, c, c.Args().First(), cb, fuzzy)
}

// see action.show - context, cli context, name, key, rescurse.
type showFunc func(context.Context, *cli.Context, string, bool) error

func (s *Action) find(ctx context.Context, c *cli.Context, needle string, cb showFunc, fuzzy bool) error {
	// get all existing entries.
	haystack, err := s.Store.List(ctx, tree.INF)
	if err != nil {
		return exit.Error(exit.List, err, "failed to list store: %s", err)
	}

	// filter our the ones from the haystack matching the needle.
	choices, err := filter(haystack, needle, c.Bool("regex"))
	if err != nil {
		return exit.Error(exit.Usage, err, "%s", err)
	}

	// if we have an exact match print it.
	if len(choices) == 1 {
		if cb == nil {
			out.Printf(ctx, choices[0])

			return nil
		}
		out.OKf(ctx, "Found exact match in %q", choices[0])

		return cb(ctx, c, choices[0], false)
	}

	// if we don't have a match yet try a fuzzy search.
	if len(choices) < 1 && fuzzy {
		// try fuzzy match.
		cm := closestmatch.New(haystack, []int{2})
		choices = cm.ClosestN(needle, 5)
	}

	// if there are still no results we abort.
	if len(choices) < 1 {
		return exit.Error(exit.NotFound, nil, "no results found")
	}

	// do not invoke wizard if not printing to terminal or if
	// gopass find/search was invoked directly (for scripts).
	if !ctxutil.IsTerminal(ctx) || (c != nil && c.Command.Name == "find") {
		for _, value := range choices {
			out.Printf(ctx, value)
		}

		return nil
	}

	return s.findSelection(ctx, c, choices, needle, cb)
}

// findSelection runs a wizard that lets the user select an entry.
func (s *Action) findSelection(ctx context.Context, c *cli.Context, choices []string, needle string, cb showFunc) error {
	if cb == nil {
		return fmt.Errorf("callback is nil")
	}
	if len(choices) < 1 {
		return fmt.Errorf("out of options")
	}

	sort.Strings(choices)
	act, sel := cui.GetSelection(ctx, "Found secrets - Please select an entry", choices)
	debug.Log("Action: %s - Selection: %d", act, sel)

	switch act {
	case "default":
		// display or copy selected entry.
		fmt.Fprintln(stdout, choices[sel])

		return cb(ctx, c, choices[sel], false)
	case "copy":
		// display selected entry.
		fmt.Fprintln(stdout, choices[sel])

		return cb(WithClip(ctx, true), c, choices[sel], false)
	case "show":
		// display selected entry.
		fmt.Fprintln(stdout, choices[sel])

		return cb(WithClip(ctx, false), c, choices[sel], false)
	case "sync":
		// run sync and re-run show/find workflow.
		if err := s.Sync(c); err != nil {
			return err
		}

		return cb(ctx, c, needle, true)
	case "edit":
		// edit selected entry.
		fmt.Fprintln(stdout, choices[sel])

		return s.edit(ctx, c, choices[sel])
	default:
		return exit.Error(exit.Aborted, nil, "user aborted")
	}
}

func filter(l []string, needle string, reMatch bool) ([]string, error) {
	choices := make([]string, 0, 10)

	if reMatch {
		compiledRE, err := regexp.Compile(needle)
		if err != nil {
			return nil, err
		}
		for _, value := range l {
			if compiledRE.MatchString(value) {
				choices = append(choices, value)
			}
		}

		return choices, nil
	}

	for _, value := range l {
		if strings.Contains(strings.ToLower(value), strings.ToLower(needle)) {
			choices = append(choices, value)
		}
	}

	return choices, nil
}
