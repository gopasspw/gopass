// Package pwgen implements multiple popular password generate algorithms.
// It supports creating classic cryptic passwords with different character
// classes as well as more recent memorable approaches.
//
// Some methods try to ensure certain requirements are met and can be very slow.
package pwgen

import (
	"fmt"
	"os"
)

const (
	digits = "0123456789"
	upper  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lower  = "abcdefghijklmnopqrstuvwxyz"
	syms   = "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"
	// CharAlpha is the class of letters
	CharAlpha = upper + lower
	// CharAlphaNum is the class of alpha-numeric characters
	CharAlphaNum = digits + upper + lower
	// CharAll is the class of all characters
	CharAll = digits + upper + lower + syms
)

// GeneratePassword generates a random, hard to remember password
func GeneratePassword(length int, symbols bool) string {
	chars := digits + upper + lower
	if symbols {
		chars += syms
	}
	if c := os.Getenv("GOPASS_CHARACTER_SET"); c != "" {
		chars = c
	}
	return GeneratePasswordCharset(length, chars)
}

// GeneratePasswordCharset generates a random password from a given
// set of characters
func GeneratePasswordCharset(length int, chars string) string {
	c := NewCryptic(length)
	c.Chars = chars
	return c.Password()
}

// GeneratePasswordWithAllClasses tries to enforce a password which
// contains all character classes instead of only enabling them.
// This is especially useful for broken (corporate) password policies
// that mandate the use of certain character classes for no good reason
func GeneratePasswordWithAllClasses(length int) (string, error) {
	c := NewCrypticWithAllClasses(length)
	if pw := c.Password(); pw != "" {
		return pw, nil
	}
	return "", fmt.Errorf("failed to generate matching password after %d rounds", c.MaxTries)
}

// GeneratePasswordCharsetCheck generates a random password from a given
// set of characters and validates the generated password with crunchy
func GeneratePasswordCharsetCheck(length int, chars string) string {
	c := NewCrypticWithCrunchy(length)
	c.Chars = chars
	return c.Password()
}
