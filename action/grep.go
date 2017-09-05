package action

import (
	"context"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/urfave/cli"
)

// Grep searches a string inside the content of all files
func (s *Action) Grep(ctx context.Context, c *cli.Context) error {
	if !c.Args().Present() {
		return s.exitError(ctx, ExitUsage, nil, "Usage: %s grep arg", s.Name)
	}

	search := c.Args().First()

	l, err := s.Store.List(0)
	if err != nil {
		return s.exitError(ctx, ExitList, err, "failed to list store: %s", err)
	}

	for _, v := range l {
		content, err := s.Store.Get(ctx, v)
		if err != nil {
			fmt.Println(color.RedString("failed to decrypt %s: %v", v, err))
			continue
		}

		if strings.Contains(string(content), search) {
			fmt.Printf("%s:\n%s", color.BlueString(v), string(content))
		}
	}

	return nil
}
