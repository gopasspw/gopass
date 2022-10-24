package pwrules

import (
	"context"
	"sort"
	"strings"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/set"
)

var customAliases map[string][]string

// LookupAliases looks up known aliases for the given domain.
func LookupAliases(ctx context.Context, domain string) []string {
	if customAliases == nil {
		_ = loadCustomAliases(ctx)
	}
	aliases := make([]string, 0, len(genAliases[domain])+len(customAliases[domain]))
	aliases = append(aliases, genAliases[domain]...)
	aliases = append(aliases, customAliases[domain]...)
	sort.Strings(aliases)

	return aliases
}

// AllAliases returns all aliases.
func AllAliases(ctx context.Context) map[string][]string {
	if customAliases == nil {
		_ = loadCustomAliases(ctx)
	}
	all := make(map[string][]string, len(genAliases)+len(customAliases))
	for k, v := range genAliases {
		all[k] = append(all[k], v...)
	}

	for k, v := range customAliases {
		all[k] = append(all[k], v...)
	}

	return all
}

func loadCustomAliases(ctx context.Context) error {
	customAliases = make(map[string][]string, 128)
	for _, k := range set.SortedFiltered(config.FromContext(ctx).Keys(""), func(k string) bool {
		return strings.HasPrefix(k, "domain-alias.")
	}) {
		from := config.String(ctx, k)
		to := strings.TrimPrefix(k, "domain-alias.")
		if e, found := customAliases[from]; found {
			e = append(e, to)
			sort.Strings(e)
			customAliases[from] = e

			continue
		}

		customAliases[from] = []string{to}
	}

	return nil
}
