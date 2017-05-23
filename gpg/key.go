package gpg

import (
	"fmt"
	"time"
)

// Key is a GPG key (public or secret)
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
}

// IsUseable returns true if GPG would assume this key is useable for encryption
func (k Key) IsUseable() bool {
	if !k.ExpirationDate.IsZero() && k.ExpirationDate.Before(time.Now()) {
		return false
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
// not exactly the same, as the output form GPG itself
func (k Key) String() string {
	fp := ""
	if len(k.Fingerprint) > 24 {
		fp = k.Fingerprint[24:]
	}
	out := fmt.Sprintf("%s   %dD/0x%s %s", k.KeyType, k.KeyLength, fp, k.CreationDate.Format("2006-01-02"))
	if !k.ExpirationDate.IsZero() {
		out += fmt.Sprintf(" [expires: %s]", k.ExpirationDate.Format("2006-01-02"))
	}
	out += "\n      Key fingerprint = " + k.Fingerprint
	for _, id := range k.Identities {
		out += fmt.Sprintf("\n" + id.String())
	}
	return out
}

// OneLine prints a terse representation of this key on one line (includes only
// the first identity!)
func (k Key) OneLine() string {
	id := Identity{}
	for _, i := range k.Identities {
		id = i
		break
	}
	return fmt.Sprintf("0x%s - %s", k.Fingerprint[24:], id.ID())
}

// ID returns the short fingerprint
func (k Key) ID() string {
	return fmt.Sprintf("0x%s", k.Fingerprint[24:])
}
