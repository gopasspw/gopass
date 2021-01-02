package otp

import (
	"io/ioutil"
	"strings"

	"github.com/gokyle/twofactor"
	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/pkg/errors"
)

// Calculate will compute a OTP code from a given secret
func Calculate(name string, sec gopass.Secret) (twofactor.OTP, string, error) {
	otpURL := ""
	// check body
	for _, line := range strings.Split(sec.Body(), "\n") {
		if strings.HasPrefix(line, "otpauth://") {
			otpURL = line
			break
		}
	}
	if otpURL != "" {
		return twofactor.FromURL(otpURL)
	}

	// check yaml entry and fall back to password if we don't have one
	label := name

	secKey, found := sec.Get("totp")
	if !found {
		secKey = sec.Password()
	}

	if strings.HasPrefix(secKey, "otpauth://") {
		return twofactor.FromURL(secKey)
	}

	otp, err := twofactor.NewGoogleTOTP(twofactor.Pad(secKey))
	return otp, label, err
}

// WriteQRFile writes the given OTP code as a QR image to disk
func WriteQRFile(otp twofactor.OTP, label, file string) error {
	var qr []byte
	var err error
	switch otp.Type() {
	case twofactor.OATH_HOTP:
		hotp := otp.(*twofactor.HOTP)
		qr, err = hotp.QR(label)
	case twofactor.OATH_TOTP:
		totp := otp.(*twofactor.TOTP)
		qr, err = totp.QR(label)
	default:
		err = errors.New("QR codes can only be generated for OATH OTPs")
	}
	if err != nil {
		return errors.Wrapf(err, "%s", err)
	}

	if err := ioutil.WriteFile(file, qr, 0600); err != nil {
		return errors.Wrapf(err, "failed to write QR code: %s", err)
	}
	return nil
}
