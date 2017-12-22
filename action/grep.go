package action

import (
	"context"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/urfave/cli"
)

// Grep searches a string inside the content of all files
func (s *Action) Grep(ctx context.Context, c *cli.Context) error {
	if !c.Args().Present() {
		return exitError(ctx, ExitUsage, nil, "Usage: %s grep arg", s.Name)
	}

	search := c.Args().First()

	l, err := s.Store.List(0)
	if err != nil {
		return exitError(ctx, ExitList, err, "failed to list store: %s", err)
	}

	for _, v := range l {
		sec, err := s.Store.Get(ctx, v)
		if err != nil {
			out.Red(ctx, "failed to decrypt %s: %v", v, err)
			continue
		}

		if strings.Contains(sec.Password(), search) {
			fmt.Printf("%s:\n%s", color.BlueString(v), sec.Password())
		}
		if strings.Contains(sec.Body(), search) {
			fmt.Printf("%s:\n%s", color.BlueString(v), sec.Body())
		}
	}

	return nil
}
