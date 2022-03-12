package otp

import (
	"fmt"
	"os"
	"strings"

	"github.com/gokyle/twofactor"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/gopass"
)

// Calculate will compute a OTP code from a given secret.
//nolint:ireturn
func Calculate(name string, sec gopass.Secret) (twofactor.OTP, string, error) {
	otpURL, found := sec.Get("otpauth")
	if found && strings.HasPrefix(otpURL, "//") {
		otpURL = "otpauth:" + otpURL
	} else {
		// check body
		for _, line := range strings.Split(sec.Body(), "\n") {
			if strings.HasPrefix(line, "otpauth://") {
				otpURL = line

				break
			}
		}
	}

	if otpURL != "" {
		debug.Log("found otpauth url: %s", out.Secret(otpURL))

		return twofactor.FromURL(otpURL) //nolint:wrapcheck
	}

	// check yaml entry and fall back to password if we don't have one
	label := name

	secKey, found := sec.Get("totp")
	if !found {
		debug.Log("no totp secret found, falling back to password")

		secKey = sec.Password()
	}

	if strings.HasPrefix(secKey, "otpauth://") {
		return twofactor.FromURL(secKey) //nolint:wrapcheck
	}

	otp, err := twofactor.NewGoogleTOTP(twofactor.Pad(secKey))
	if err != nil {
		return otp, label, fmt.Errorf("invalid OTP secret %q: %w", secKey, err)
	}

	return otp, label, nil
}

// WriteQRFile writes the given OTP code as a QR image to disk.
func WriteQRFile(otp twofactor.OTP, label, file string) error {
	var qr []byte

	var err error

	switch otp.Type() {
	case twofactor.OATH_HOTP:
		hotp, ok := otp.(*twofactor.HOTP)
		if !ok {
			return fmt.Errorf("Type assertion failed on twofactor.HOTP: %w", ErrType)
		}

		qr, err = hotp.QR(label)
	case twofactor.OATH_TOTP:
		totp, ok := otp.(*twofactor.TOTP)
		if !ok {
			return fmt.Errorf("Type assertion failed on twofactor.TOTP: %w", ErrType)
		}

		qr, err = totp.QR(label)
	default:
		err = ErrOathOTP
	}

	if err != nil {
		return fmt.Errorf("failed to write qr file: %w", err)
	}

	if err := os.WriteFile(file, qr, 0o600); err != nil {
		return fmt.Errorf("failed to write QR code: %w", err)
	}

	return nil
}

var (
	// ErrOathOTP is returned when the secret is not a valid OATH secret.
	ErrOathOTP = fmt.Errorf("QR codes can only be generated for OATH OTPs")
	// ErrType is returned when the secret is not a valid OTP type.
	ErrType = fmt.Errorf("type assertion failed")
)
