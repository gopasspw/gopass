package action

import (
	"context"
	"sort"
	"strings"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/pwgen/pwrules"
	"github.com/urfave/cli/v3"
)

// AliasesPrint prints all configured aliases for password generation rules.
func (s *miscHandler) AliasesPrint(ctx context.Context, cmd *cli.Command) error {
	out.Printf(ctx, "Configured aliases:")
	aliases := pwrules.AllAliases(ctx)
	keys := make([]string, 0, len(aliases))
	for k := range aliases {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	for _, k := range keys {
		out.Printf(ctx, "- %s -> %s", k, strings.Join(aliases[k], ", "))
	}

	return nil
}
