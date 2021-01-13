package bcrypt

import "golang.org/x/crypto/bcrypt"

const (
	cost = 12
)

// Generate generates a new Bcrypt hash with recommended values for it's
// cost parameter.
func Generate(password string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}

	return "{BLF-CRYPT}" + string(h), nil
}
