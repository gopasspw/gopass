package argon2id

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

var (
	// ErrInvalidHash is returned if the required parameters can not be obtained
	// from the hash.
	ErrInvalidHash = fmt.Errorf("argon2id: invalid hash format")

	// ErrVersionIncompatible is returned if the Argon2id version generating
	// the hash does not match the version validating it.
	ErrVersionIncompatible = fmt.Errorf("argon2id: incompatible version")

	// Prefix is set to be compatible with Dovecot. Can be set to an empty string.
	Prefix = "{ARGON2ID}"
)

// DefaultParams provides sane default parameters for password hashing as of
// 2021. Depending on your environment you will need to adjust these.
var DefaultParams = &Params{
	Memory:      512 * 1024,
	Iterations:  3,
	Parallelism: 4,
	SaltLen:     32,
	KeyLen:      32,
}

// Params contains the input parameters for the Argon2id algorithm. Memory and
// Iterations tweak the computational cost. If you have more cores available
// you can change the parallelism to reduce runtime without reducing cost. But
// note that this will change the hash.
//
// See https://tools.ietf.org/html/draft-irtf-cfrg-argon2-04#section-4
type Params struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLen     uint32
	KeyLen      uint32
}

// Generate generates a new Argon2ID hash with recommended values for it's
// complexity parameters. By default the generated hash is compatible with
// the Dovecot Password Scheme.
//
// See https://doc.dovecot.org/configuration_manual/authentication/password_schemes/
//
// It looks like this
//
//	{ARGON2ID}$argon2id$v=19$m=524288,t=3,p=4$464unwkIcBGXjqWBZ0A5FWClURgYdWFqRlQaBJOE5fs$5ofdht4OkXsg/tftXGgxNchAdgHzpe+QJyizabiKZFk
func Generate(password string, saltLen uint32) (string, error) {
	params := DefaultParams
	if saltLen > 0 {
		params.SaltLen = saltLen
	}

	salt := make([]byte, params.SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to read rand: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, params.Iterations, params.Memory, params.Parallelism, params.KeyLen)

	b64hash := base64.RawStdEncoding.EncodeToString(hash)
	b64salt := base64.RawStdEncoding.EncodeToString(salt)

	return fmt.Sprintf(Prefix+"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, params.Memory, params.Iterations, params.Parallelism, b64salt, b64hash), nil
}

// Validate unpacks the parameters from the hash, computes the hash of the given
// password with these parameters and performs a constant time comparison between
// both hashes.
func Validate(password string, hash string) (bool, error) {
	params, salt, key, err := unpackHash(hash)
	if err != nil {
		return false, err
	}

	otherKey := argon2.IDKey([]byte(password), salt, params.Iterations, params.Memory, params.Parallelism, params.KeyLen)

	if subtle.ConstantTimeEq(int32(len(key)), int32(len(otherKey))) == 0 {
		return false, nil
	}

	if subtle.ConstantTimeCompare(key, otherKey) == 1 {
		return true, nil
	}

	return false, nil
}

func unpackHash(hash string) (*Params, []byte, []byte, error) {
	hash = strings.TrimPrefix(hash, Prefix)

	p := strings.Split(hash, "$")
	if len(p) != 6 {
		return nil, nil, nil, ErrInvalidHash
	}

	if p[1] != "argon2id" {
		return nil, nil, nil, ErrInvalidHash
	}

	var version int

	_, err := fmt.Sscanf(p[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to scan version %s: %w", p[2], err)
	}

	if version != argon2.Version {
		return nil, nil, nil, ErrVersionIncompatible
	}

	params := &Params{}

	_, err = fmt.Sscanf(p[3], "m=%d,t=%d,p=%d", &params.Memory, &params.Iterations, &params.Parallelism)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to scan header %s: %w", p[3], err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(p[4])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to decode salt %s: %w", p[4], err)
	}

	params.SaltLen = uint32(len(salt))

	key, err := base64.RawStdEncoding.DecodeString(p[5])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to decode hash %s: %w", p[5], err)
	}

	params.KeyLen = uint32(len(key))

	return params, salt, key, nil
}
