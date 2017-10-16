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
	if !c.Args().Present() {
		return s.exitError(ctx, ExitUsage, nil, "Usage: %s find arg", s.Name)
	}

	l, err := s.Store.List(0)
	if err != nil {
		return s.exitError(ctx, ExitList, err, "failed to list store: %s", err)
	}

	needle := strings.ToLower(c.Args().First())
	choices := make([]string, 0, 10)
	for _, value := range l {
		if strings.Contains(strings.ToLower(value), needle) {
			choices = append(choices, value)
		}
	}

	clip := c.Bool("clip")

	if len(choices) == 1 {
		out.Green(ctx, "Found exact match in '%s'", choices[0])
		return s.show(ctx, c, choices[0], "", clip, false, false, false)
	}

	if len(choices) < 1 {
		// try fuzzy match
		cm := closestmatch.New(l, []int{2})
		choices = cm.ClosestN(needle, 5)
		if len(choices) < 1 {
			return fmt.Errorf("no results found")
		}
	}

	if !ctxutil.IsTerminal(ctx) || c.Command.Name == "find" {
		for _, value := range choices {
			fmt.Println(value)
		}
		return nil
	}

	act, sel := termwiz.GetSelection(ctx, "Found secrets -", "", choices)
	switch act {
	case "copy":
		// display selected entry
		fmt.Println(choices[sel])
		return s.show(ctx, c, choices[sel], "", true, false, false, false)
	case "show":
		// display selected entry
		fmt.Println(choices[sel])
		return s.show(ctx, c, choices[sel], "", clip, false, false, false)
	case "sync":
		// run sync and re-run show/find workflow
		if err := s.Sync(ctx, c); err != nil {
			return err
		}
		return s.show(ctx, c, needle, "", clip, false, false, true)
	default:
		return s.exitError(ctx, ExitAborted, nil, "user aborted")
	}
}
