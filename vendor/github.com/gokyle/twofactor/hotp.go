package twofactor

import (
	"crypto"
	"crypto/sha1"
	"encoding/base32"
	"io"
	"net/url"
	"strconv"
	"strings"
)

// HOTP represents an RFC-4226 Hash-based One Time Password instance.
type HOTP struct {
	*OATH
}

// Type returns OATH_HOTP.
func (otp *HOTP) Type() Type {
	return OATH_HOTP
}

// NewHOTP takes the key, the initial counter value, and the number
// of digits (typically 6 or 8) and returns a new HOTP instance.
func NewHOTP(key []byte, counter uint64, digits int) *HOTP {
	return &HOTP{
		OATH: &OATH{
			key:     key,
			counter: counter,
			size:    digits,
			hash:    sha1.New,
			algo:    crypto.SHA1,
		},
	}
}

// OTP returns the next OTP and increments the counter.
func (otp *HOTP) OTP() string {
	code := otp.OATH.OTP(otp.counter)
	otp.counter++
	return code
}

// URL returns an HOTP URL (i.e. for putting in a QR code).
func (otp *HOTP) URL(label string) string {
	return otp.OATH.URL(otp.Type(), label)
}

// SetProvider sets up the provider component of the OTP URL.
func (otp *HOTP) SetProvider(provider string) {
	otp.provider = provider
}

// GenerateGoogleHOTP generates a new HOTP instance as used by
// Google Authenticator.
func GenerateGoogleHOTP() *HOTP {
	key := make([]byte, sha1.Size)
	if _, err := io.ReadFull(PRNG, key); err != nil {
		return nil
	}
	return NewHOTP(key, 0, 6)
}

func hotpFromURL(u *url.URL) (*HOTP, string, error) {
	label := u.Path[1:]
	v := u.Query()

	secret := strings.ToUpper(v.Get("secret"))
	if secret == "" {
		return nil, "", ErrInvalidURL
	}

	var digits = 6
	if sdigit := v.Get("digits"); sdigit != "" {
		tmpDigits, err := strconv.ParseInt(sdigit, 10, 8)
		if err != nil {
			return nil, "", err
		}
		digits = int(tmpDigits)
	}

	var counter uint64 = 0
	if scounter := v.Get("counter"); scounter != "" {
		var err error
		counter, err = strconv.ParseUint(scounter, 10, 64)
		if err != nil {
			return nil, "", err
		}
	}

	key, err := base32.StdEncoding.DecodeString(Pad(secret))
	if err != nil {
		// assume secret isn't base32 encoded
		key = []byte(secret)
	}
	otp := NewHOTP(key, counter, digits)
	return otp, label, nil
}

// QR generates a new QR code for the HOTP.
func (otp *HOTP) QR(label string) ([]byte, error) {
	return otp.OATH.QR(otp.Type(), label)
}
