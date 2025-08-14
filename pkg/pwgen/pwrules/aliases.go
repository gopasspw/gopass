package pwrules

import (
	"context"
	"sort"
	"strings"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/set"
)

// LookupAliases looks up known aliases for the given domain.
func LookupAliases(ctx context.Context, domain string) []string {
	customAliases := loadCustomAliases(ctx)
	aliases := make([]string, 0, len(genAliases[domain])+len(customAliases[domain]))
	aliases = append(aliases, genAliases[domain]...)
	aliases = append(aliases, customAliases[domain]...)
	sort.Strings(aliases)

	return aliases
}

// AllAliases returns all aliases.
func AllAliases(ctx context.Context) map[string][]string {
	customAliases := loadCustomAliases(ctx)
	all := make(map[string][]string, len(genAliases)+len(customAliases))
	for k, v := range genAliases {
		all[k] = append(all[k], v...)
	}

	for k, v := range customAliases {
		all[k] = append(all[k], v...)
	}

	return all
}

func loadCustomAliases(ctx context.Context) map[string][]string {
	cfg, _ := config.FromContext(ctx)
	customAliases := make(map[string][]string, 128)
	for _, k := range set.SortedFiltered(cfg.Keys(""), func(k string) bool {
		return strings.HasPrefix(k, "domain-alias.") && strings.HasSuffix(k, ".insteadof")
	}) {
		// NB: we currently only support system, env, global or local <root> store level aliases
		for _, from := range cfg.GetAll(k) {
			to := strings.TrimSuffix(strings.TrimPrefix(k, "domain-alias."), ".insteadof")
			debug.Log("Loading alias: %q -> %q", from, to)
			if e, found := customAliases[from]; found {
				e = append(e, to)
				sort.Strings(e)
				customAliases[from] = e

				continue
			}

			customAliases[from] = []string{to}
		}
	}

	return customAliases
}
