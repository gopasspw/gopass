package action

import (
	"sort"
	"strings"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/pwgen/pwrules"
	"github.com/urfave/cli/v2"
)

// AliasesPrint prints all cofigured aliases
func (s *Action) AliasesPrint(c *cli.Context) error {
	out.Print(c.Context, "Configured aliases:")
	aliases := pwrules.AllAliases()
	keys := make([]string, 0, len(aliases))
	for k := range aliases {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		out.Print(c.Context, "- %s -> %s", k, strings.Join(aliases[k], ", "))
	}
	return nil
}

// AliasesAdd adds a single alias to a domain
func (s *Action) AliasesAdd(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	domain := c.Args().First()
	alias := c.Args().Get(1)

	if domain == "" || alias == "" {
		return ExitError(ExitUsage, nil, "Usage: %s alias add <domain> <alias>", s.Name)
	}

	if err := pwrules.AddCustomAlias(domain, alias); err != nil {
		return err
	}

	out.Print(ctx, "Added alias '%s' to domain '%s'", alias, domain)
	return nil
}

// AliasesRemove removes a single alias from a domain
func (s *Action) AliasesRemove(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	domain := c.Args().First()
	alias := c.Args().Get(1)

	if domain == "" || alias == "" {
		return ExitError(ExitUsage, nil, "Usage: %s alias remove <domain> <alias>", s.Name)
	}

	if err := pwrules.RemoveCustomAlias(domain, alias); err != nil {
		return err
	}

	out.Print(ctx, "Remove alias '%s' from domain '%s'", alias, domain)
	return nil
}

// AliasesDelete remove an alias mapping for a domain
func (s *Action) AliasesDelete(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	domain := c.Args().First()

	if domain == "" {
		return ExitError(ExitUsage, nil, "Usage: %s alias delete <domain>", s.Name)
	}

	if err := pwrules.DeleteCustomAlias(domain); err != nil {
		return err
	}

	out.Print(ctx, "Remove aliases for domain '%s'", domain)
	return nil
}
