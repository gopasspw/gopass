package age

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/gopasspw/gopass/pkg/debug"
)

// FindIdentities it TODO
func (a *Age) FindIdentities(ctx context.Context, keys ...string) ([]string, error) {
	nk, err := a.getAllIdentities(ctx)
	if err != nil {
		return nil, err
	}
	matches := make([]string, 0, len(nk))
	for _, k := range keys {
		debug.Log("Key: %s", k)
		if _, found := nk[k]; found {
			debug.Log("Found")
			matches = append(matches, k)
			continue
		}
		debug.Log("not found in %+v", nk)
	}
	sort.Strings(matches)
	return matches, nil
}

// FindRecipients is TODO
func (a *Age) FindRecipients(ctx context.Context, keys ...string) ([]string, error) {
	// TODO should not need to decrypt keyring
	remote := make([]string, 0, len(keys))
	local := make([]string, 0, len(keys))
	for _, key := range keys {
		if !strings.HasPrefix(key, "github:") {
			local = append(local, key)
			continue
		}
		pks, err := a.getPublicKeysGithub(ctx, strings.TrimPrefix(key, "github:"))
		if err != nil {
			debug.Log("Failed to get key %s from github: %s", key, err)
			continue
		}
		remote = append(remote, pks...)
	}
	ids, err := a.FindIdentities(ctx, local...)
	if err != nil {
		return nil, err
	}
	return append(ids, remote...), nil
}

// FormatKey is TODO
func (a *Age) FormatKey(ctx context.Context, id, tpl string) string {
	return id
}

// Fingerprint return the id
func (a *Age) Fingerprint(ctx context.Context, id string) string {
	return id
}

// ListRecipients is TODO
func (a *Age) ListRecipients(context.Context) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}

// ReadNamesFromKey is TODO
func (a *Age) ReadNamesFromKey(ctx context.Context, buf []byte) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}

// RecipientIDs is not supported by design
func (a *Age) RecipientIDs(ctx context.Context, buf []byte) ([]string, error) {
	return nil, fmt.Errorf("reading recipient IDs is not supported by the age backend by design")
}
