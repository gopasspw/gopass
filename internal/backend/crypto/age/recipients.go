package age

import (
	"context"
	"strings"

	"filippo.io/age"
	"filippo.io/age/agessh"
	"github.com/gopasspw/gopass/pkg/debug"
)

// FindRecipients returns all list of usable recipient key IDs matching the search strings.
// For native age keys this is a no-op since they are self-contained (i.e. the ID is the full key already).
// But for SSH keys, especially GitHub indirections, an extra step is necessary.
func (a *Age) FindRecipients(ctx context.Context, search ...string) ([]string, error) {
	remote := make([]string, 0, len(search))
	local := make([]string, 0, len(search))
	for _, key := range search {
		if !strings.HasPrefix(key, "github:") {
			local = append(local, key)
			continue
		}
		pks, err := a.ghCache.ListKeys(ctx, strings.TrimPrefix(key, "github:"))
		if err != nil {
			debug.Log("Failed to get key %s from github: %s", key, err)
			continue
		}
		remote = append(remote, pks...)
	}
	recps, err := a.IdentityRecipients(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(remote)+len(local)+len(recps))
	for _, r := range recipientsToBech32(recps) {
		for _, l := range local {
			if r == l {
				ids = append(ids, r)
			}
		}
	}
	recp := append(ids, remote...)
	debug.Log("found usable keys for %q: %q (all: %q)", search, recp, append(ids, remote...))
	return recp, nil
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
