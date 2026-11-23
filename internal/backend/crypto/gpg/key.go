package gpg

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// SubKey represents a GPG subkey parsed from --with-colons output.
type SubKey struct {
	KeyID          string
	Fingerprint    string
	Validity       string
	CreationDate   time.Time
	ExpirationDate time.Time
	Caps           Capabilities
}

// IsUseable returns true if GPG would assume this subkey is useable for encryption.
func (sk SubKey) IsUseable(alwaysTrust bool) bool {
	if sk.Caps.Deactivated {
		return false
	}

	if !sk.Caps.Encrypt {
		return false
	}

	if !sk.ExpirationDate.IsZero() && sk.ExpirationDate.Before(time.Now()) {
		return false
	}

	// Expired, revoked, disabled, or invalid subkeys cannot be encrypted to
	// even when alwaysTrust is enabled.
	switch sk.Validity {
	case "e", "r", "d", "i":
		return false
	}

	if alwaysTrust {
		return true
	}

	switch sk.Validity {
	case "m", "f", "u":
		return true
	}

	return false
}

// Key is a GPG key (public or secret).
type Key struct {
	KeyType        string
	KeyLength      int
	Validity       string
	CreationDate   time.Time
	ExpirationDate time.Time
	Ownertrust     string
	Fingerprint    string
	Identities     map[string]Identity
	SubKeys        map[string]struct{}
	SubKeyList     []SubKey
	Caps           Capabilities
}

// Capabilities of a Key.
type Capabilities struct {
	Encrypt        bool
	Sign           bool
	Certify        bool
	Authentication bool
	Deactivated    bool
}

// IsUseable returns true if GPG would assume this key is useable for encryption.
func (k Key) IsUseable(alwaysTrust bool) bool {
	if k.Caps.Deactivated {
		return false
	}

	if !k.Caps.Encrypt {
		return false
	}

	if !k.ExpirationDate.IsZero() && k.ExpirationDate.Before(time.Now()) {
		return false
	}

	if alwaysTrust {
		return true
	}

	switch k.Validity {
	case "m":
		return true
	case "f":
		return true
	case "u":
		return true
	}

	return false
}

// String implement fmt.Stringer. This method produces output that is close to, but
// not exactly the same, as the output from GPG itself.
func (k Key) String() string {
	var out strings.Builder
	fmt.Fprintf(&out, "%s   %dD/%s %s", k.KeyType, k.KeyLength, k.ID(), k.CreationDate.Format("2006-01-02"))
	if !k.ExpirationDate.IsZero() {
		fmt.Fprintf(&out, " [expires: %s]", k.ExpirationDate.Format("2006-01-02"))
	}

	out.WriteString("\n      Key fingerprint = " + k.Fingerprint)
	for _, id := range k.Identities {
		fmt.Fprintf(&out, "\n%s", id)
	}

	return out.String()
}

// OneLine prints a terse representation of this key on one line (includes only
// the first identity!).
func (k Key) OneLine() string {
	id := k.ID()
	if id == "" {
		return fmt.Sprintf("(invalid:%s)", k.Fingerprint)
	}

	return fmt.Sprintf("%s - %s", id, k.Identity().ID())
}

// Identity returns the first identity.
func (k Key) Identity() Identity {
	ids := make([]Identity, 0, len(k.Identities))
	for _, i := range k.Identities {
		ids = append(ids, i)
	}

	sort.Slice(ids, func(i, j int) bool {
		if !ids[i].CreationDate.Equal(ids[j].CreationDate) {
			return ids[i].CreationDate.After(ids[j].CreationDate)
		}

		return ids[i].UID < ids[j].UID
	})

	for _, i := range ids {
		return i
	}

	return Identity{}
}

// ID returns the short fingerprint.
func (k Key) ID() string {
	if len(k.Fingerprint) == 64 {
		return fmt.Sprintf("0x%s", k.Fingerprint[:16])
	}
	if len(k.Fingerprint) >= 40 {
		return fmt.Sprintf("0x%s", k.Fingerprint[len(k.Fingerprint)-16:])
	}
	if len(k.Fingerprint) >= 24 {
		return fmt.Sprintf("0x%s", k.Fingerprint[24:])
	}

	return ""
}

// MatchSubKey finds a subkey matching the given key ID or fingerprint
// (with optional "0x"/"0X" prefix and/or trailing "!").
func (k Key) MatchSubKey(id string) (SubKey, bool) {
	hexID := strings.TrimSpace(id)
	hexID = strings.TrimSuffix(hexID, "!")
	if strings.HasPrefix(hexID, "0x") || strings.HasPrefix(hexID, "0X") {
		hexID = hexID[2:]
	}
	hexID = strings.ToUpper(hexID)
	if hexID == "" {
		return SubKey{}, false
	}

	for _, sk := range k.SubKeyList {
		fp := strings.ToUpper(sk.Fingerprint)
		kid := strings.ToUpper(strings.TrimPrefix(strings.TrimPrefix(sk.KeyID, "0x"), "0X"))
		if (fp != "" && (fp == hexID || strings.HasSuffix(fp, hexID))) ||
			(kid != "" && (kid == hexID || strings.HasSuffix(kid, hexID))) {
			return sk, true
		}
	}

	// Fallback for synthetic Key values that only populate SubKeys.
	if len(k.SubKeyList) == 0 {
		for sk := range k.SubKeys {
			sub := strings.ToUpper(strings.TrimPrefix(strings.TrimPrefix(sk, "0x"), "0X"))
			if sub == hexID || strings.HasSuffix(sub, hexID) {
				return SubKey{
					KeyID:          sub,
					Fingerprint:    sub,
					Validity:       k.Validity,
					CreationDate:   k.CreationDate,
					ExpirationDate: k.ExpirationDate,
					Caps:           k.Caps,
				}, true
			}
		}
	}

	return SubKey{}, false
}
