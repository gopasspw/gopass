package bcrypt

import (
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const (
	cost = 12
)

// Prefix is set to be compatible with Dovecot. Can be set to an empty string.
var Prefix = "{BLF-CRYPT}"

// Generate generates a new Bcrypt hash with recommended values for it's
// cost parameter.
func Generate(password string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", fmt.Errorf("failed to generate password hash: %w", err)
	}

	return Prefix + string(h), nil
}

// Validate validates the password against the given hash.
func Validate(password, hash string) error {
	hash = strings.TrimPrefix(hash, Prefix)

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return fmt.Errorf("failed to validate password hash %s: %w", hash, err)
	}

	return nil
}
