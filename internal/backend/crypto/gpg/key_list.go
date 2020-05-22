package gpg

import (
	"sort"
	"strings"

	"github.com/pkg/errors"
)

// KeyList is a searchable slice of Keys
type KeyList []Key

// Recipients returns the KeyList formatted as a recipient list
func (kl KeyList) Recipients() []string {
	l := make([]string, 0, len(kl))
	sort.Sort(kl)
	for _, k := range kl {
		l = append(l, k.ID())
	}
	return l
}

// UseableKeys returns the list of useable (valid keys)
func (kl KeyList) UseableKeys() KeyList {
	nkl := make(KeyList, 0, len(kl))
	sort.Sort(kl)
	for _, k := range kl {
		if !k.IsUseable() {
			continue
		}
		nkl = append(nkl, k)
	}
	return nkl
}

// UnusableKeys returns the list of unusable keys (invalid keys)
func (kl KeyList) UnusableKeys() KeyList {
	nkl := make(KeyList, 0, len(kl))
	for _, k := range kl {
		if k.IsUseable() {
			continue
		}
		nkl = append(nkl, k)
	}
	sort.Sort(nkl)
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
			if ident.Name == id {
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
	return Key{}, errors.Errorf("No matching key found")
}

func (kl KeyList) Len() int {
	return len(kl)
}

func (kl KeyList) Less(i, j int) bool {
	return kl[i].Identity().Name < kl[j].Identity().Name
}

func (kl KeyList) Swap(i, j int) {
	kl[i], kl[j] = kl[j], kl[i]
}
