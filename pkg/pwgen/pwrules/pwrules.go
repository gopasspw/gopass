package pwrules

import (
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/gopasspw/gopass/pkg/debug"
)

//go:generate go run gen.go

var reChars = regexp.MustCompile(`(allowed|required):\s*\[(.*)\](?:;|,)`)

// AllRules returns all rules.
func AllRules() map[string]Rule {
	return genRules
}

// LookupRule looks up a rule either directly or through one of it's know
// aliases.
func LookupRule(domain string) (Rule, bool) {
	r, found := genRules[domain]
	if found {
		return r, true
	}

	for _, alias := range LookupAliases(domain) {
		if r, found := genRules[alias]; found {
			return r, true
		}
	}

	return Rule{}, false
}

// Rule is a password rule as defined by Apple at https://developer.apple.com/password-rules/
type Rule struct {
	Minlen    int
	Maxlen    int
	Required  []string
	Allowed   []string
	Maxconsec int
	Exact     bool
}

// ParseRule parses a password rule.
// NOTE: This is not a complete parser.
func ParseRule(in string) Rule {
	r := Rule{}

	if reChars.MatchString(in) {
		m := reChars.FindStringSubmatch(in)
		if len(m) > 2 {
			re := "[" + m[2] + "]"

			switch m[1] {
			case "required":
				r.Required = append(r.Required, re)
			case "allowed":
				r.Allowed = append(r.Allowed, re)
			}
		}
	}

	for _, part := range strings.Split(strings.TrimSuffix(in, ";"), ";") {
		p := strings.Split(part, ": ")
		if len(p) < 2 {
			continue
		}

		var err error

		key := strings.TrimSpace(p[0])
		strVal := strings.TrimSpace(p[1])
		max := len(strVal)

		if i := strings.Index(strVal, "["); i > 0 {
			max = i
		}

		switch key {
		case "minlength":
			r.Minlen, err = strconv.Atoi(strVal)
		case "maxlength":
			r.Maxlen, err = strconv.Atoi(strVal)
		case "max-consecutive":
			r.Maxconsec, err = strconv.Atoi(strVal)
		case "required":
			r.Required = append(r.Required, strings.Split(strVal[0:max], ",")...)
		case "allowed":
			r.Allowed = append(r.Allowed, strings.Split(strVal[0:max], ",")...)
		}

		if err != nil {
			debug.Log("failed to parse %s for %s: %s", strVal, key, err)
		}
	}

	r.Required = sanitize(r.Required)
	r.Allowed = sanitize(r.Allowed)

	return r
}

func sanitize(in []string) []string {
	out := make([]string, 0, len(in))

	for _, v := range in {
		v := strings.TrimSpace(v)
		if strings.HasPrefix(v, "[") && !strings.HasSuffix(v, "]") {
			continue
		}

		out = append(out, v)
	}

	sort.Strings(out)

	return out
}
