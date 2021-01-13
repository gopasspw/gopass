package pwrules

import (
	"math"
	"strconv"
	"strings"

	"github.com/gopasspw/gopass/pkg/debug"
)

//go:generate go run gen.go

var (
	rules = map[string]Rule{}
)

func init() {
	for k, v := range genRules {
		// do not override customizations
		if _, found := rules[k]; found {
			continue
		}
		r := ParseRule(v)
		r.Exact = genRulesExact[k]
		if r.Maxlen < 1 {
			r.Maxlen = math.MaxInt32
		}
		rules[k] = r
	}
}

// AllRules returns all rules
func AllRules() map[string]Rule {
	return rules
}

// LookupRule looks up a rule either directly or through one of it's know
// aliases.
func LookupRule(domain string) (Rule, bool) {
	r, found := rules[domain]
	if found {
		return r, true
	}
	for _, alias := range LookupAliases(domain) {
		if r, found := rules[alias]; found {
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
	for _, part := range strings.Split(strings.TrimSuffix(in, ";"), "; ") {
		p := strings.Split(part, ": ")
		if len(p) < 2 {
			continue
		}
		var err error
		key := p[0]
		strVal := p[1]
		switch key {
		case "minlength":
			r.Minlen, err = strconv.Atoi(strVal)
		case "maxlength":
			r.Maxlen, err = strconv.Atoi(strVal)
		case "max-consecutive":
			r.Maxconsec, err = strconv.Atoi(strVal)
		case "required":
			r.Required = append(r.Required, strings.Split(strVal, ", ")...)
		case "allowed":
			r.Allowed = append(r.Allowed, strings.Split(strVal, ", ")...)
		}
		if err != nil {
			debug.Log("failed to parse %s for %s: %s", strVal, key, err)
		}
	}
	return r
}
