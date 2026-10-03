// Package identityorder preserves gopass identity preferences during age decryption.
package identityorder

import "filippo.io/age"

// identity delegates Unwrap without exposing the underlying concrete type.
// age prioritizes native identity types before other identities, which would
// otherwise override an explicitly preferred plugin identity.
type identity struct {
	age.Identity
}

// Preserve returns identities that age tries in their supplied order.
// Apply this only at the age.Decrypt boundary: callers still need the original
// identity types for recipient lookup and serialization to the agent.
func Preserve(ids []age.Identity) []age.Identity {
	out := make([]age.Identity, len(ids))
	for i, id := range ids {
		out[i] = identity{Identity: id}
	}

	return out
}
