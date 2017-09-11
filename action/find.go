package action

import (
	"context"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/utils/ctxutil"
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

	if len(choices) == 1 {
		fmt.Println(color.GreenString("Found exact match in '%s'", choices[0]))
		return s.show(ctx, c, choices[0], "", false, false, false)
	}

	if len(choices) < 1 {
		// try fuzzy match
		cm := closestmatch.New(l, []int{2})
		choices = cm.ClosestN(needle, 5)
		if len(choices) < 1 {
			return fmt.Errorf("no results found")
		}
	}

	if !ctxutil.IsTerminal(ctx) {
		for _, value := range choices {
			fmt.Println(value)
		}
		return nil
	}

	act, sel := termwiz.GetSelection(choices)
	switch act {
	case "copy":
		return s.show(ctx, c, choices[sel], "", true, false, false)
	case "show":
		return s.show(ctx, c, choices[sel], "", false, false, false)
	default:
		return s.exitError(ctx, ExitAborted, nil, "user aborted")
	}
}
