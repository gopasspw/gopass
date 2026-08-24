package age

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"filippo.io/age"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/pkg/debug"
)

// orderedIdentities returns the given identities in a stable and
// user-controllable order (see https://github.com/gopasspw/gopass/issues/3393).
//
// Go map iteration is randomized, so without this the identity age tries
// first for decryption would be picked at random, leading to unpredictable
// prompts (e.g. sometimes asking for a hardware token PIN, sometimes not).
//
// The order is:
//  1. identities whose recipient is listed in the age.identities config
//     option, in that order,
//  2. all remaining identities, sorted deterministically (native age
//     identities first, then by their string encoding).
//
// Only public recipient strings (age1...) are ever matched against or
// stored in the config, never secret key material.
func orderedIdentities(ctx context.Context, ids map[string]age.Identity) []age.Identity {
	preferred := preferredIdentities(ctx)

	// map each identity to its public recipient string, so we can match
	// and order by non-secret identifiers only.
	recpOf := make(map[string]string, len(ids)) // identity key -> recipient
	byRecp := make(map[string]string, len(ids)) // recipient -> identity key
	for k, id := range ids {
		r := recipientOf(id)
		if r == "" {
			continue
		}
		recpOf[k] = r
		if _, dup := byRecp[r]; !dup {
			byRecp[r] = k
		}
	}

	out := make([]age.Identity, 0, len(ids))
	used := make(map[string]struct{}, len(ids))

	// 1. preferred identities, in the configured order.
	for _, want := range preferred {
		if k, found := byRecp[want]; found {
			out = append(out, ids[k])
			used[k] = struct{}{}

			continue
		}
		debug.Log("preferred identity %s not found among the %d available identities", want, len(ids))
	}

	// 2. all remaining identities, deterministically sorted.
	rest := make([]string, 0, len(ids))
	for k := range ids {
		if _, done := used[k]; done {
			continue
		}
		rest = append(rest, k)
	}
	slices.SortFunc(rest, func(a, b string) int {
		// order by the public recipient where available, falling back
		// to the map key for identities without a recipient.
		ra, rb := recpOf[a], recpOf[b]
		if ra == "" {
			ra = a
		}
		if rb == "" {
			rb = b
		}

		return identitySortFunc(ra, rb)
	})

	for _, k := range rest {
		out = append(out, ids[k])
	}

	debug.Log("ordered %d identities, %d of them preferred", len(out), len(used))

	return out
}

// recipientOf returns the public recipient string of the given identity,
// or an empty string if it cannot be determined. It never returns secret
// key material.
func recipientOf(id age.Identity) string {
	if rec := IdentityToRecipient(id); rec != nil {
		if s, ok := rec.(fmt.Stringer); ok {
			return s.String()
		}
	}

	return ""
}

// identitySortFunc orders recipient strings deterministically: native age
// recipients (age1...) first, then everything else alphabetically. This
// mirrors the ordering applied to recipients in encrypt.go so that the
// local, most likely passwordless identity is tried before hardware tokens
// that may require a PIN or touch.
func identitySortFunc(a, b string) int {
	if la, lb := len(a), len(b); la != lb {
		// yubikey and other plugin identities are typically longer
		return la - lb
	}

	return strings.Compare(a, b)
}

// preferredIdentities returns the user's preferred identity order from the
// age.identities config option. Entries may be given either as a single
// comma-separated value or as repeated config entries. Only public
// recipient strings (age1...) should be stored there.
func preferredIdentities(ctx context.Context) []string {
	prefs := config.Strings(ctx, "age.identities")

	if len(prefs) < 1 {
		return nil
	}

	// split comma-separated values and dedupe, preserving order.
	out := make([]string, 0, len(prefs))
	seen := make(map[string]struct{}, len(prefs))
	for _, p := range prefs {
		for _, e := range strings.Split(p, ",") {
			e = strings.TrimSpace(e)
			if e == "" {
				continue
			}
			if _, dup := seen[e]; dup {
				continue
			}
			seen[e] = struct{}{}
			out = append(out, e)
		}
	}

	return out
}
