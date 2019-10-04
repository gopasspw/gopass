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

	var matches int
	var errors int
	for _, v := range haystack {
		sec, err := s.Store.Get(ctx, v)
		if err != nil {
			out.Error(ctx, "failed to decrypt %s: %v", v, err)
			errors++
			continue
		}

		if strings.Contains(sec.Password(), needle) {
			out.Print(ctx, "%s:\n%s", color.BlueString(v), sec.Password())
			matches++
		}
		if strings.Contains(sec.Body(), needle) {
			out.Print(ctx, "%s:\n%s", color.BlueString(v), sec.Body())
			matches++
		}
	}

	if errors > 0 {
		return ExitError(ctx, ExitDecrypt, nil, "some secrets failed to decrypt")
	}
	if matches < 1 {
		return ExitError(ctx, ExitNotFound, nil, "no matches found")
	}

	out.Print(ctx, "\nScanned %d secrets. %d matches, %d errors", len(haystack), matches, errors)
	return nil
}
