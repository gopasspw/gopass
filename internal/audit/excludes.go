package audit

import (
	"regexp"
	"strings"

	"github.com/gopasspw/gopass/pkg/debug"
)

type res []*regexp.Regexp

func (r res) Matches(s string) bool {
	for _, re := range r {
		if re.MatchString(s) {
			debug.Log("Matched %s against %s", s, re.String())

			return true
		}
	}

	return false
}

// FilterExcludes filters the given list of secrets against the given exclude patterns (RE2 syntax).
func FilterExcludes(excludes string, in []string) []string {
	debug.Log("Filtering %d secrets against %d exclude patterns", len(in), strings.Count(excludes, "\n"))

	res := make(res, 0, 10)
	for _, line := range strings.Split(excludes, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		re, err := regexp.Compile(line)
		if err != nil {
			debug.Log("failed to compile exclude pattern %q: %s", line, err)

			continue
		}
		debug.Log("Adding exclude pattern %q", re.String())
		res = append(res, re)
	}

	// shortcut if we have no excludes
	if len(res) < 1 {
		return in
	}

	// check all secrets against all excludes
	out := make([]string, 0, len(in))
	for _, s := range in {
		if res.Matches(s) {
			continue
		}
		out = append(out, s)
	}

	return out
}
