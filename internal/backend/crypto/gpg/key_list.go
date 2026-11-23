package gpg

import (
	"fmt"
	"sort"
	"strings"
)

// ErrKeyNotFound is returned when a key is not found.
var ErrKeyNotFound = fmt.Errorf("no matching key found")

// KeyList is a searchable slice of Keys.
type KeyList []Key

// Recipients returns the KeyList formatted as a recipient list.
func (kl KeyList) Recipients() []string {
	sort.Sort(kl)
	u := make(map[string]struct{}, len(kl))

	for _, k := range kl {
		u[k.ID()] = struct{}{}
	}

	l := make([]string, 0, len(kl))
	for k := range u {
		l = append(l, k)
	}

	sort.Strings(l)

	return l
}

// UseableKeys returns the list of useable (valid keys).
func (kl KeyList) UseableKeys(alwaysTrust bool) KeyList {
	nkl := make(KeyList, 0, len(kl))
	sort.Sort(kl)

	for _, k := range kl {
		if !k.IsUseable(alwaysTrust) {
			continue
		}

		nkl = append(nkl, k)
	}

	return nkl
}

// UnusableKeys returns the list of unusable keys (invalid keys).
func (kl KeyList) UnusableKeys(alwaysTrust bool) KeyList {
	nkl := make(KeyList, 0, len(kl))

	for _, k := range kl {
		if k.IsUseable(alwaysTrust) {
			continue
		}

		nkl = append(nkl, k)
	}

	sort.Sort(nkl)

	return nkl
}

// FindKey will try to find the requested key.
func (kl KeyList) FindKey(id string) (Key, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Key{}, ErrKeyNotFound
	}

	hexID := strings.TrimSuffix(id, "!")
	if strings.HasPrefix(hexID, "0x") || strings.HasPrefix(hexID, "0X") {
		hexID = hexID[2:]
	}
	hexID = strings.ToUpper(hexID)

	emailID := strings.TrimPrefix(strings.TrimSuffix(id, ">"), "<")

	for _, k := range kl {
		if hexID != "" {
			fp := strings.ToUpper(k.Fingerprint)
			if fp == hexID || strings.HasSuffix(fp, hexID) || strings.HasPrefix(fp, hexID) {
				return k, nil
			}
		}

		for _, ident := range k.Identities {
			if ident.Name == id {
				return k, nil
			}

			if emailID != "" && strings.EqualFold(ident.Email, emailID) {
				return k, nil
			}
		}

		if hexID != "" {
			for sk := range k.SubKeys {
				sub := strings.ToUpper(strings.TrimPrefix(strings.TrimPrefix(sk, "0x"), "0X"))
				if strings.HasSuffix(sub, hexID) || strings.HasPrefix(sub, hexID) {
					return k, nil
				}
			}
		}
	}

	return Key{}, ErrKeyNotFound
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
