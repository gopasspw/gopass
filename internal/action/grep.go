package action

import (
	"strings"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

// Grep searches a string inside the content of all files
func (s *Action) Grep(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	if !c.Args().Present() {
		return ExitError(ExitUsage, nil, "Usage: %s grep arg", s.Name)
	}

	// get the search term
	needle := c.Args().First()

	haystack, err := s.Store.List(ctx, 0)
	if err != nil {
		return ExitError(ExitList, err, "failed to list store: %s", err)
	}

	var matches int
	var errors int
	for _, v := range haystack {
		sec, err := s.Store.Get(ctx, v)
		if err != nil {
			out.Error(ctx, "failed to decrypt %s: %v", v, err)
			continue
		}

		if strings.Contains(string(sec.Bytes()), needle) {
			out.Print(ctx, "%s matches", color.BlueString(v))
		}
	}

	out.Red(ctx, "WARNING: some secrets failed to decrypt")
	out.Print(ctx, "\nScanned %d secrets. %d matches, %d errors", len(haystack), matches, errors)
	return nil
}
