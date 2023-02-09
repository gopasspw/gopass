package age

import (
	"context"
	"strings"

	"filippo.io/age"
	"filippo.io/age/agessh"
	"github.com/gopasspw/gopass/internal/set"
	"github.com/gopasspw/gopass/pkg/debug"
)

// FindRecipients returns all list of usable recipient key IDs matching the search strings.
// For native age keys this is a no-op since they are self-contained (i.e. the ID is the full key already).
// But for SSH keys, especially GitHub indirections, an extra step is necessary.
func (a *Age) FindRecipients(ctx context.Context, search ...string) ([]string, error) {
	rs := set.New[string]()

	for _, key := range search {
		switch {
		case strings.HasPrefix(key, "github:"):
			// look up any "github:<username>" style public SSH keys
			pks, err := a.ghCache.ListKeys(ctx, strings.TrimPrefix(key, "github:"))
			if err != nil {
				debug.Log("Failed to get key %s from github: %s", key, err)

				continue
			}

			rs.Add(pks...)
		case strings.HasPrefix(key, "ssh-"):
			// add ssh public keys as-is
			rs.Add(key)
		case strings.HasPrefix(key, "age1"):
			// add any regular age public keys as-is
			rs.Add(key)
		default:
			debug.Log("ignoring unknown key: %s", key)
		}
	}

	debug.Log("found usable keys for %q: %q ", search, rs)

	return rs.Elements(), nil
}

func (a *Age) parseRecipients(ctx context.Context, recipients []string) ([]age.Recipient, error) {
	out := make([]age.Recipient, 0, len(recipients))
	for _, r := range recipients {
		if strings.HasPrefix(r, "age1") {
			id, err := age.ParseX25519Recipient(r)
			if err != nil {
				debug.Log("Failed to parse recipient %q as X25519: %s", r, err)

				continue
			}
			out = append(out, id)

			continue
		}
		if strings.HasPrefix(r, "ssh-") {
			id, err := agessh.ParseRecipient(r)
			if err != nil {
				debug.Log("Failed to parse recipient %q as SSH: %s", r, err)

				continue
			}
			out = append(out, id)

			continue
		}
		if strings.HasPrefix(r, "github:") {
			pks, err := a.ghCache.ListKeys(ctx, strings.TrimPrefix(r, "github:"))
			if err != nil {
				return out, err
			}
			for _, pk := range pks {
				id, err := agessh.ParseRecipient(r)
				if err != nil {
					debug.Log("Failed to parse GitHub recipient %q: %q: %s", r, pk, err)

					continue
				}
				out = append(out, id)
			}
		}
	}

	return out, nil
}
