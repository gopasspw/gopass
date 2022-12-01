package action

import (
	"sort"
	"strings"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/pwgen/pwrules"
	"github.com/urfave/cli/v2"
)

// AliasesPrint prints all cofigured aliases.
func (s *Action) AliasesPrint(c *cli.Context) error {
	out.Printf(c.Context, "Configured aliases:")
	aliases := pwrules.AllAliases(c.Context)
	keys := make([]string, 0, len(aliases))
	for k := range aliases {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	for _, k := range keys {
		out.Printf(c.Context, "- %s -> %s", k, strings.Join(aliases[k], ", "))
	}

	return nil
}
