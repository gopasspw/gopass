// Package pwgen implements multiple popular password generate algorithms.
// It supports creating classic cryptic passwords with different character
// classes as well as more recent memorable approaches.
//
// Some methods try to ensure certain requirements are met and can be very slow.
package pwgen

import (
	"fmt"
	"os"
	"strings"
)

// Character classes
const (
	Digits = "0123456789"
	Upper  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	Lower  = "abcdefghijklmnopqrstuvwxyz"
	Syms   = "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"
	Ambiq  = "0ODQ1IlB8G6S5Z2"
	// CharAlpha is the class of letters
	CharAlpha = Upper + Lower
	// CharAlphaNum is the class of alpha-numeric characters
	CharAlphaNum = Digits + Upper + Lower
	// CharAll is the class of all characters
	CharAll = Digits + Upper + Lower + Syms
)

// GeneratePassword generates a random, hard to remember password
func GeneratePassword(length int, symbols bool) string {
	chars := Digits + Upper + Lower
	if symbols {
		chars += Syms
	}
	if c := os.Getenv("GOPASS_CHARACTER_SET"); c != "" {
		chars = c
	}
	return GeneratePasswordCharset(length, chars)
}

// GeneratePasswordCharset generates a random password from a given
// set of characters
func GeneratePasswordCharset(length int, chars string) string {
	c := NewCryptic(length, false)
	c.Chars = chars
	return c.Password()
}

// GeneratePasswordWithAllClasses tries to enforce a password which
// contains all character classes instead of only enabling them.
// This is especially useful for broken (corporate) password policies
// that mandate the use of certain character classes for no good reason
func GeneratePasswordWithAllClasses(length int, symbols bool) (string, error) {
	c := NewCrypticWithAllClasses(length, symbols)
	if pw := c.Password(); pw != "" {
		return pw, nil
	}
	return "", fmt.Errorf("failed to generate matching password after %d rounds", c.MaxTries)
}

// GeneratePasswordCharsetCheck generates a random password from a given
// set of characters and validates the generated password with crunchy
func GeneratePasswordCharsetCheck(length int, chars string) string {
	c := NewCrypticWithCrunchy(length, false)
	c.Chars = chars
	return c.Password()
}

// Prune removes all characters in cutset from the input
func Prune(in string, cutset string) string {
	out := make([]rune, 0, len(in))
	for _, r := range in {
		if strings.Contains(cutset, string(r)) {
			continue
		}
		out = append(out, r)
	}
	return string(out)
}
