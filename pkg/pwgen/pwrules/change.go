package pwrules

import "context"

var changeURLs = map[string]string{}

func init() {
	for k, v := range genChange {
		// filter out invalid entries
		if v == "" {
			continue
		}

		changeURLs[k] = v
	}
}

// LookupChangeURL looks up a change URL, either directly or through
// one of its known aliases.
func LookupChangeURL(ctx context.Context, domain string) string {
	if u, found := changeURLs[domain]; found {
		return u
	}

	for _, alias := range LookupAliases(ctx, domain) {
		if u, found := changeURLs[alias]; found {
			return u
		}
	}

	return ""
}
