package gpg

import (
	"fmt"
	"strings"
)

// KeyList is a searchable slice of Keys
type KeyList []Key

// UseableKeys returns the list of useable (valid keys)
func (kl KeyList) UseableKeys() KeyList {
	nkl := make(KeyList, 0, len(kl))
	for _, k := range kl {
		if !k.IsUseable() {
			continue
		}
		nkl = append(nkl, k)
	}
	return nkl
}

// FindKey will try to find the requested key
func (kl KeyList) FindKey(id string) (Key, error) {
	id = strings.TrimPrefix(id, "0x")
	for _, k := range kl {
		if k.Fingerprint == id {
			return k, nil
		}
		if strings.HasSuffix(k.Fingerprint, id) {
			return k, nil
		}
		for _, ident := range k.Identities {
			if ident.String() == id {
				return k, nil
			}
			if ident.Email == id {
				return k, nil
			}
		}
		for sk := range k.SubKeys {
			if strings.HasSuffix(sk, id) {
				return k, nil
			}
		}
	}
	return Key{}, fmt.Errorf("No matching key found")
}
