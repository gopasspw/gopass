package pwgen

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/pwgen/pwrules"
	"github.com/muesli/crunchy"
)

// Cryptic is a generator for hard-to-remember passwords as required by (too)
// many sites. Prefer memorable or xkcd-style passwords, if possible.
type Cryptic struct {
	Chars      string
	Length     int
	MaxTries   int
	Validators []func(string) error
}

// NewCryptic creates a new generator with sane defaults.
func NewCryptic(length int, symbols bool) *Cryptic {
	if length < 1 {
		length = 16
	}
	chars := Digits + Upper + Lower
	if symbols {
		chars += Syms
	}
	return &Cryptic{
		Chars:    chars,
		Length:   length,
		MaxTries: 64,
	}
}

// NewCrypticForDomain tries to look up password rules for the given domain
// or uses the default generator.
func NewCrypticForDomain(length int, domain string) *Cryptic {
	c := NewCryptic(length, true)
	r, found := pwrules.LookupRule(domain)
	debug.Log("found rules for %s: %t", domain, found)
	if !found {
		return c
	}
	if r.Maxlen > 0 && c.Length > r.Maxlen {
		c.Length = r.Maxlen
	}
	if c.Length < r.Minlen {
		c.Length = r.Minlen
	}
	if chars := charsFromRule(append(r.Required, r.Allowed...)...); chars != "" {
		c.Chars = chars
	}
	for _, req := range r.Required {
		chars := charsFromRule(req)
		if req == "" || strings.TrimSpace(chars) == "" {
			continue
		}
		debug.Log("Adding validator for %s: Requires %q -> %q", domain, req, chars)
		c.Validators = append(c.Validators, func(pw string) error {
			wantChars := charsFromRule(req)
			if wantChars == "" {
				return nil
			}
			if containsAllClasses(pw, wantChars) {
				return nil
			}
			return fmt.Errorf("password %s does not contain any of %s", pw, chars)
		})
	}
	if r.Maxconsec > 0 {
		c.Validators = append(c.Validators, func(pw string) error {
			if containsMaxConsecutive(pw, r.Maxconsec) {
				return nil
			}
			return fmt.Errorf("password %s contains more than %d consecutive characters", pw, r.Maxconsec)
		})
	}
	debug.Log("initialized generator: %+v", c)
	return c
}

func charsFromRule(rules ...string) string {
	chars := ""
	for _, req := range rules {
		switch req {
		case "lower":
			chars += Lower
		case "upper":
			chars += Upper
		case "digit":
			chars += Digits
		case "special":
			chars += Syms
		default:
			if strings.HasPrefix(req, "[") && strings.HasSuffix(req, "]") {
				chars += strings.Trim(req, "[]")
			}
		}
	}
	return uniqueChars(chars)
}

func uniqueChars(in string) string {
	// a set of chars, not a charset
	charSet := make(map[rune]struct{}, len(in))
	for _, c := range in {
		charSet[c] = struct{}{}
	}
	charSlice := make([]string, 0, len(charSet))
	for k := range charSet {
		charSlice = append(charSlice, string(k))
	}
	sort.Strings(charSlice)
	return strings.Join(charSlice, "")
}

// NewCrypticWithAllClasses returns a password generator that generates passwords
// containing all available character classes
func NewCrypticWithAllClasses(length int, symbols bool) *Cryptic {
	c := NewCryptic(length, symbols)
	c.Validators = append(c.Validators, func(pw string) error {
		if containsAllClasses(pw, c.Chars) {
			return nil
		}
		return fmt.Errorf("password does not contain all character classes")
	})
	return c
}

// NewCrypticWithCrunchy returns a password generators that only returns a
// password if it's successfully validated with crunchy.
func NewCrypticWithCrunchy(length int, symbols bool) *Cryptic {
	c := NewCryptic(length, symbols)
	c.MaxTries = 3
	validator := crunchy.NewValidator()
	c.Validators = append(c.Validators, validator.Check)
	return c
}

// Password returns a single password from the generator
func (c *Cryptic) Password() string {
	round := 0
	maxFn := func() bool {
		round++
		if c.MaxTries < 1 {
			return false
		}
		if c.MaxTries == 0 && round >= 64 {
			return true
		}
		if round > c.MaxTries {
			return true
		}
		return false
	}
	for {
		if maxFn() {
			debug.Log("failed to generate password after %d rounds", round)
			return ""
		}
		if pw := c.randomString(); c.isValid(pw) {
			return pw
		}
	}
}

func (c *Cryptic) isValid(pw string) bool {
	for _, v := range c.Validators {
		if err := v(pw); err != nil {
			debug.Log("failed to validate: %s", err)
			return false
		}
	}
	return true
}

func (c *Cryptic) randomString() string {
	pw := &bytes.Buffer{}
	for pw.Len() < c.Length {
		_ = pw.WriteByte(c.Chars[randomInteger(len(c.Chars))])
	}
	return pw.String()
}
