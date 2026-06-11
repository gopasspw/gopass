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
	"github.com/urfave/cli/v3"
)

// Find runs find without fuzzy search.
func (s *searchHandler) Find(ctx context.Context, cmd *cli.Command) error {
	return s.findCmd(ctx, cmd, nil, false)
}

// FindFuzzy runs find with fuzzy search.
func (s *searchHandler) FindFuzzy(ctx context.Context, cmd *cli.Command) error {
	return s.findCmd(ctx, cmd, s.showFn, true)
}

func (s *searchHandler) findCmd(ctx context.Context, cmd *cli.Command, cb showFunc, fuzzy bool) error {
	ctx = ctxutil.WithGlobalFlags(ctx, cmd)
	// Note: do not re-parse the clip flag here. The find command has no clip flag,
	// and when called via fuzzy search from show, the context already carries the
	// correct clip state set by showParseArgs. Calling cmd.Bool("clip") on the show
	// command's GenericFlag (OptionalInt) returns false in cli v3, which would
	// incorrectly overwrite the clip=true already stored in ctx.

	if cmd.IsSet("unsafe") {
		ctx = ctxutil.WithForce(ctx, cmd.Bool("unsafe"))
	}

	if !cmd.Args().Present() {
		return exit.Error(exit.Usage, nil, "Usage: %s find <pattern>", s.Name)
	}

	return s.find(ctx, cmd, cmd.Args().First(), cb, fuzzy)
}

// see action.show - context, cli context, name, key, rescurse.
type showFunc func(context.Context, *cli.Command, string, bool) error

func (s *searchHandler) find(ctx context.Context, cmd *cli.Command, needle string, cb showFunc, fuzzy bool) error {
	// get all existing entries.
	haystack, err := s.Store.List(ctx, tree.INF)
	if err != nil {
		return exit.Error(exit.List, err, "failed to list store: %s", err)
	}

	// filter our the ones from the haystack matching the needle.
	choices, err := filter(haystack, needle, cmd.Bool("regex"))
	if err != nil {
		return exit.Error(exit.Usage, err, "%s", err)
	}

	// if we don't have a match yet try a fuzzy search.
	if len(choices) < 1 && fuzzy {
		// try fuzzy match.
		cm := closestmatch.New(haystack, []int{2})
		choices = cm.ClosestN(needle, 5)
	}

	// JSON output: emit the matches as a JSON array (possibly empty) and return.
	if cmd != nil && cmd.Bool("json") {
		if len(choices) < 1 {
			return exit.Error(exit.NotFound, nil, "no results found")
		}

		return jsonWrite(stdout, choices)
	}

	// if we have an exact match print it.
	if len(choices) == 1 {
		if cb == nil {
			out.Printf(ctx, choices[0])

			return nil
		}
		out.OKf(ctx, "Found exact match in %q", choices[0])

		return cb(ctx, cmd, choices[0], false)
	}

	// if there are still no results we abort.
	if len(choices) < 1 {
		return exit.Error(exit.NotFound, nil, "no results found")
	}

	// do not invoke wizard if not printing to terminal or if
	// gopass find/search was invoked directly (for scripts).
	if !ctxutil.IsTerminal(ctx) || (cmd != nil && cmd.Name == "find") {
		for _, value := range choices {
			out.Printf(ctx, value)
		}

		return nil
	}

	return s.findSelection(ctx, cmd, choices, needle, cb)
}

// findSelection runs a wizard that lets the user select an entry.
func (s *searchHandler) findSelection(ctx context.Context, cmd *cli.Command, choices []string, needle string, cb showFunc) error {
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

		return cb(ctx, cmd, choices[sel], false)
	case "copy":
		// display selected entry.
		fmt.Fprintln(stdout, choices[sel])

		return cb(WithClip(ctx, true), cmd, choices[sel], false)
	case "show":
		// display selected entry.
		fmt.Fprintln(stdout, choices[sel])

		return cb(WithClip(ctx, false), cmd, choices[sel], false)
	case "sync":
		// run sync and re-run show/find workflow.
		if err := s.syncFn(ctx, cmd); err != nil {
			return err
		}

		return cb(ctx, cmd, needle, true)
	case "edit":
		// edit selected entry.
		fmt.Fprintln(stdout, choices[sel])

		return s.editFn(ctx, cmd, choices[sel])
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
