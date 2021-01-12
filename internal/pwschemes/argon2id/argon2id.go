package argon2id

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/argon2"
)

const (
	time    uint32 = 3
	memory  uint32 = 512 * 1024
	threads uint8  = 4
	keylen  uint32 = 32
)

// Generate generates a new Argon2ID hash with recommended values for it's
// complexity parameters.
func Generate(password string, saltLen uint8) (string, error) {
	salt := make([]byte, saltLen)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, time, memory, threads, keylen)

	hstr := base64.RawStdEncoding.EncodeToString(hash)
	sstr := base64.RawStdEncoding.EncodeToString(salt)

	return fmt.Sprintf("{ARGON2ID}$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, memory, time, threads, sstr, hstr), nil
}
