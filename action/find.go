package action

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/termwiz"
	"github.com/urfave/cli"
)

// Find a string in the secret file's name
func (s *Action) Find(c *cli.Context) error {
	if !c.Args().Present() {
		return s.exitError(ExitUsage, nil, "Usage: %s find arg", s.Name)
	}

	l, err := s.Store.List(0)
	if err != nil {
		return s.exitError(ExitList, err, "failed to list store: %s", err)
	}

	needle := strings.ToLower(c.Args().First())
	choices := make([]string, 0, 10)
	for _, value := range l {
		if strings.Contains(strings.ToLower(value), needle) {
			choices = append(choices, value)
		}
	}

	if len(choices) < 1 {
		return fmt.Errorf("no results found")
	}

	if len(choices) == 1 {
		fmt.Println(color.GreenString("Found exact match in '%s'", choices[0]))
		return s.show(c, choices[0], "", false, false, false)
	}

	if !s.isTerm {
		for _, value := range choices {
			fmt.Println(value)
		}
		return nil
	}

	act, sel := termwiz.GetSelection(choices)
	switch act {
	case "copy":
		return s.show(c, choices[sel], "", true, false, false)
	case "show":
		return s.show(c, choices[sel], "", false, false, false)
	default:
		return s.exitError(ExitAborted, nil, "user aborted")
	}
}
