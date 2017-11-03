package twofactor

import (
	"crypto"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base32"
	"hash"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// TOTP represents an RFC 6238 Time-based One-Time Password instance.
type TOTP struct {
	*OATH
	step uint64
}

// Type returns OATH_TOTP.
func (otp *TOTP) Type() Type {
	return OATH_TOTP
}

func (otp *TOTP) otp(counter uint64) string {
	return otp.OATH.OTP(counter)
}

// OTP returns the OTP for the current timestep.
func (otp *TOTP) OTP() string {
	return otp.otp(otp.OTPCounter())
}

// URL returns a TOTP URL (i.e. for putting in a QR code).
func (otp *TOTP) URL(label string) string {
	return otp.OATH.URL(otp.Type(), label)
}

// SetProvider sets up the provider component of the OTP URL.
func (otp *TOTP) SetProvider(provider string) {
	otp.provider = provider
}

func (otp *TOTP) otpCounter(t uint64) uint64 {
	return (t - otp.counter) / otp.step
}

// OTPCounter returns the current time value for the OTP.
func (otp *TOTP) OTPCounter() uint64 {
	return otp.otpCounter(uint64(time.Now().Unix()))
}

// NewOTP takes a new key, a starting time, a step, the number of
// digits of output (typically 6 or 8) and the hash algorithm to
// use, and builds a new OTP.
func NewTOTP(key []byte, start uint64, step uint64, digits int, algo crypto.Hash) *TOTP {
	h := hashFromAlgo(algo)
	if h == nil {
		return nil
	}

	return &TOTP{
		OATH: &OATH{
			key:     key,
			counter: start,
			size:    digits,
			hash:    h,
			algo:    algo,
		},
		step: step,
	}

}

// NewTOTPSHA1 will build a new TOTP using SHA-1.
func NewTOTPSHA1(key []byte, start uint64, step uint64, digits int) *TOTP {
	return NewTOTP(key, start, step, digits, crypto.SHA1)
}

func hashFromAlgo(algo crypto.Hash) func() hash.Hash {
	switch algo {
	case crypto.SHA1:
		return sha1.New
	case crypto.SHA256:
		return sha256.New
	case crypto.SHA512:
		return sha512.New
	}
	return nil
}

// GenerateGoogleTOTP produces a new TOTP token with the defaults expected by
// Google Authenticator.
func GenerateGoogleTOTP() *TOTP {
	key := make([]byte, sha1.Size)
	if _, err := io.ReadFull(PRNG, key); err != nil {
		return nil
	}
	return NewTOTP(key, 0, 30, 6, crypto.SHA1)
}

// NewGoogleTOTP takes a secret as a base32-encoded string and
// returns an appropriate Google Authenticator TOTP instance.
func NewGoogleTOTP(secret string) (*TOTP, error) {
	key, err := base32.StdEncoding.DecodeString(secret)
	if err != nil {
		return nil, err
	}
	return NewTOTP(key, 0, 30, 6, crypto.SHA1), nil
}

func totpFromURL(u *url.URL) (*TOTP, string, error) {
	label := u.Path[1:]
	v := u.Query()

	secret := strings.ToUpper(v.Get("secret"))
	if secret == "" {
		return nil, "", ErrInvalidURL
	}

	var algo = crypto.SHA1
	if algorithm := v.Get("algorithm"); algorithm != "" {
		switch {
		case algorithm == "SHA256":
			algo = crypto.SHA256
		case algorithm == "SHA512":
			algo = crypto.SHA512
		case algorithm != "SHA1":
			return nil, "", ErrInvalidAlgo
		}
	}

	var digits = 6
	if sdigit := v.Get("digits"); sdigit != "" {
		tmpDigits, err := strconv.ParseInt(sdigit, 10, 8)
		if err != nil {
			return nil, "", err
		}
		digits = int(tmpDigits)
	}

	var period uint64 = 30
	if speriod := v.Get("period"); speriod != "" {
		var err error
		period, err = strconv.ParseUint(speriod, 10, 64)
		if err != nil {
			return nil, "", err
		}
	}

	key, err := base32.StdEncoding.DecodeString(secret)
	if err != nil {
		return nil, "", err
	}
	otp := NewTOTP(key, 0, period, digits, algo)
	return otp, label, nil
}

// QR generates a new TOTP QR code.
func (otp *TOTP) QR(label string) ([]byte, error) {
	return otp.OATH.QR(otp.Type(), label)
}
