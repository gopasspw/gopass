// Package secparse provides functions to parse secrets from various formats.
// It can parse secrets from legacy MIME format, YAML format, and AKV format.
package secparse

import (
	"errors"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
)

// Parse tries to parse a secret from a byte slice. It attempts to parse the
// secret in the following order: legacy MIME, YAML, and finally AKV.
// If parsing as legacy MIME or YAML fails, it falls back to the next format.
// If a permanent error is encountered while parsing as legacy MIME, it returns
// the error immediately.
//
//nolint:ireturn
func Parse(in []byte) (gopass.Secret, error) {
	var s gopass.Secret

	var err error

	s, err = parseLegacyMIME(in)
	if err == nil {
		debug.Log("parsed as MIME: %+v", s)

		return s, nil
	}

	debug.Log("failed to parse as MIME: %s", out.Secret(err.Error()))

	var permError *secrets.PermanentError
	if errors.As(err, &permError) {
		return secrets.ParseAKV(in), err
	}

	s, err = secrets.ParseYAML(in)
	if err == nil {
		debug.Log("parsed as YAML: %+v", s)

		return s, nil
	}

	debug.Log("failed to parse as YAML: %s\n%s", err, out.Secret(string(in)))

	s = secrets.ParseAKV(in)
	debug.Log("parsed as AVK: %+v", s)

	return s, nil
}

// MustParse parses a secret from a string or panics if an error occurs.
// This function should only be used for testing purposes.
func MustParse(in string) gopass.Secret {
	sec, err := Parse([]byte(in))
	if err != nil {
		panic(err)
	}

	return sec
}
