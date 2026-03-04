package action

import (
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/gopasspw/gopass/internal/action/exit"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/tree"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/urfave/cli/v2"
)

// Grep searches a string inside the content of all files.
func (s *Action) Grep(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	if !c.Args().Present() {
		return exit.Error(exit.Usage, nil, "Usage: %s grep arg", s.Name)
	}

	// get the search term.
	needle := c.Args().First()

	haystack, err := s.Store.List(ctx, tree.INF)
	if err != nil {
		return exit.Error(exit.List, err, "failed to list store: %s", err)
	}

	matchFn := func(haystack string) bool {
		return strings.Contains(haystack, needle)
	}

	if c.Bool("regexp") {
		re, err := regexp.Compile(needle)
		if err != nil {
			return exit.Error(exit.Usage, err, "failed to compile regexp %q: %s", needle, err)
		}
		matchFn = re.MatchString
	}

	var matches int
	var errors int
	for _, v := range haystack {
		sec, err := s.Store.Get(ctx, v)
		if err != nil {
			out.Errorf(ctx, "failed to decrypt %s: %v", v, err)

			continue
		}

		if matchFn(string(sec.Bytes())) {
			out.Printf(ctx, "%s matches", color.BlueString(v))
		}
	}

	if errors > 0 {
		out.Warningf(ctx, "%d secrets failed to decrypt", errors)
	}
	out.Printf(ctx, "\nScanned %d secrets. %d matches, %d errors", len(haystack), matches, errors)

	return nil
}
