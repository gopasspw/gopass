package bcrypt

import (
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const (
	cost = 12
)

var (
	// Prefix is set to be compatible with Dovecot. Can be set to an empty string.
	Prefix = "{BLF-CRYPT}"
)

// Generate generates a new Bcrypt hash with recommended values for it's
// cost parameter.
func Generate(password string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}

	return Prefix + string(h), nil
}

// Validate validates the password against the given hash.
func Validate(password, hash string) error {
	hash = strings.TrimPrefix(hash, Prefix)
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
