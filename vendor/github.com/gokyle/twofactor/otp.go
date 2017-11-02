package twofactor

import (
	"crypto/rand"
	"errors"
	"fmt"
	"hash"
	"net/url"
)

type Type uint

const (
	OATH_HOTP = iota
	OATH_TOTP
)

// PRNG is an io.Reader that provides a cryptographically secure
// random byte stream.
var PRNG = rand.Reader

var (
	ErrInvalidURL  = errors.New("twofactor: invalid URL")
	ErrInvalidAlgo = errors.New("twofactor: invalid algorithm")
)

// Type OTP represents a one-time password token -- whether a
// software taken (as in the case of Google Authenticator) or a
// hardware token (as in the case of a YubiKey).
type OTP interface {
	// Returns the current counter value; the meaning of the
	// returned value is algorithm-specific.
	Counter() uint64

	// Set the counter to a specific value.
	SetCounter(uint64)

	// the secret key contained in the OTP
	Key() []byte

	// generate a new OTP
	OTP() string

	// the output size of the OTP
	Size() int

	// the hash function used by the OTP
	Hash() func() hash.Hash

	// Returns the type of this OTP.
	Type() Type
}

func otpString(otp OTP) string {
	var typeName string
	switch otp.Type() {
	case OATH_HOTP:
		typeName = "OATH-HOTP"
	case OATH_TOTP:
		typeName = "OATH-TOTP"
	default:
		typeName = "UNKNOWN"
	}
	return fmt.Sprintf("%s, %d", typeName, otp.Size())
}

// FromURL constructs a new OTP token from a URL string.
func FromURL(URL string) (OTP, string, error) {
	u, err := url.Parse(URL)
	if err != nil {
		return nil, "", err
	}

	if u.Scheme != "otpauth" {
		return nil, "", ErrInvalidURL
	}

	switch {
	case u.Host == "totp":
		return totpFromURL(u)
	case u.Host == "hotp":
		return hotpFromURL(u)
	default:
		return nil, "", ErrInvalidURL
	}
}
