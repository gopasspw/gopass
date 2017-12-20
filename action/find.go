package action

import (
	"context"
	"fmt"
	"strings"

	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/justwatchcom/gopass/utils/termwiz"
	"github.com/schollz/closestmatch"
	"github.com/urfave/cli"
)

// Find a string in the secret file's name
func (s *Action) Find(ctx context.Context, c *cli.Context) error {
	ctx = WithClip(ctx, c.Bool("clip"))

	if !c.Args().Present() {
		return exitError(ctx, ExitUsage, nil, "Usage: %s find arg", s.Name)
	}

	l, err := s.Store.List(0)
	if err != nil {
		return exitError(ctx, ExitList, err, "failed to list store: %s", err)
	}

	needle := strings.ToLower(c.Args().First())
	choices := filter(l, needle)

	// if we have an exact match print it
	if len(choices) == 1 {
		out.Green(ctx, "Found exact match in '%s'", choices[0])
		return s.show(ctx, c, choices[0], "", false)
	}

	// if we don't have a match yet try a fuzzy search
	if len(choices) < 1 && ctxutil.IsFuzzySearch(ctx) {
		// try fuzzy match
		cm := closestmatch.New(l, []int{2})
		choices = cm.ClosestN(needle, 5)
	}

	// if there are still no results we abort
	if len(choices) < 1 {
		return fmt.Errorf("no results found")
	}

	// do not invoke wizard if not printing to terminal or if
	// gopass find/search was invoked directly (for scripts)
	if !ctxutil.IsTerminal(ctx) || c.Command.Name == "find" {
		for _, value := range choices {
			fmt.Println(value)
		}
		return nil
	}

	return s.findSelection(ctx, c, choices, needle)
}

func (s *Action) findSelection(ctx context.Context, c *cli.Context, choices []string, needle string) error {
	act, sel := termwiz.GetSelection(ctx, "Found secrets -", "", choices)
	switch act {
	case "default":
		// display or copy selected entry
		fmt.Println(choices[sel])
		return s.show(ctx, c, choices[sel], "", false)
	case "copy":
		// display selected entry
		fmt.Println(choices[sel])
		return s.show(WithClip(ctx, true), c, choices[sel], "", false)
	case "show":
		// display selected entry
		fmt.Println(choices[sel])
		return s.show(WithClip(ctx, false), c, choices[sel], "", false)
	case "sync":
		// run sync and re-run show/find workflow
		if err := s.Sync(ctx, c); err != nil {
			return err
		}
		return s.show(ctx, c, needle, "", true)
	default:
		return exitError(ctx, ExitAborted, nil, "user aborted")
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
