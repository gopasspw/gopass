package gitconfig

import "strings"

// splitKey splits a fully qualified gitconfig key into two or three parts.
// A valid key consists of either a section and a key separated by a dot
// or section, subsection and key, all separated by a dot. Note that
// the subsection might contain dots itself.
//
// Valid examples:
// - core.push
// - insteadof.git@github.com.push.
func splitKey(key string) (section, subsection, skey string) { //nolint:nonamedreturns
	n := strings.Index(key, ".")
	if n > 0 {
		section = key[:n]
	}

	if m := strings.LastIndex(key, "."); n != m && m > 0 && len(key) > m+1 {
		subsection = key[n+1 : m]
		skey = key[m+1:]

		return
	}

	skey = key[n+1:]

	return
}

func trim(s []string) {
	for i, e := range s {
		s[i] = strings.TrimSpace(e)
	}
}
