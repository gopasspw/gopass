package action

import (
	"context"
	"strings"

	"github.com/gopasspw/gopass/pkg/out"

	"github.com/fatih/color"
	"github.com/urfave/cli"
)

// Grep searches a string inside the content of all files
func (s *Action) Grep(ctx context.Context, c *cli.Context) error {
	if !c.Args().Present() {
		return ExitError(ctx, ExitUsage, nil, "Usage: %s grep arg", s.Name)
	}

	// get the search term
	needle := c.Args().First()

	haystack, err := s.Store.List(ctx, 0)
	if err != nil {
		return ExitError(ctx, ExitList, err, "failed to list store: %s", err)
	}

	for _, v := range haystack {
		sec, err := s.Store.Get(ctx, v)
		if err != nil {
			out.Red(ctx, "failed to decrypt %s: %v", v, err)
			continue
		}

		if strings.Contains(sec.Password(), needle) {
			out.Print(ctx, "%s:\n%s", color.BlueString(v), sec.Password())
		}
		if strings.Contains(sec.Body(), needle) {
			out.Print(ctx, "%s:\n%s", color.BlueString(v), sec.Body())
		}
	}

	return nil
}
