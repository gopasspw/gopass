package action

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gokyle/twofactor"
	"github.com/urfave/cli"
)

const (
	// TODO - replace this with the currently un-exported step value
	// from twofactor.FromURL
	otpPeriod = 30
)

// OTP implements OTP token handling for TOTP and HOTP
func (s *Action) OTP(ctx context.Context, c *cli.Context) error {
	name := c.Args().First()
	if name == "" {
		return s.exitError(ctx, ExitUsage, nil, "usage: %s otp [name]", s.Name)
	}

	sec, err := s.Store.Get(ctx, name)
	if err != nil {
		return s.exitError(ctx, ExitDecrypt, err, "failed to get entry '%s': %s", name, err)
	}

	otpURL := ""
	for _, line := range strings.Split(sec.Body(), "\n") {
		if strings.HasPrefix(line, "otpauth://") {
			otpURL = line
			break
		}
	}

	var otp twofactor.OTP
	var label string

	if otpURL == "" {
		// check yaml entry and fall back to password if we don't have one
		label = name
		secKey, err := sec.Value("totp")
		if err != nil {
			secKey = sec.Password()
		}

		otp, err = twofactor.NewGoogleTOTP(secKey)
		if err != nil {
			return s.exitError(ctx, ExitUnknown, err, "No OTP entry found for %s", name)
		}
	} else {
		otp, label, err = twofactor.FromURL(otpURL)
		if err != nil {
			return s.exitError(ctx, ExitUnknown, err, "failed get key from URL: %s", err)
		}
	}

	token := otp.OTP()

	now := time.Now()
	t := now.Add(otpPeriod * time.Second)

	expiresAt := time.Unix(t.Unix()+otpPeriod-(t.Unix()%otpPeriod), 0)
	secondsLeft := int(time.Until(expiresAt).Seconds())

	if secondsLeft >= otpPeriod {
		secondsLeft = secondsLeft - otpPeriod
	}

	color.Yellow("%s lasts %ds \t|%s%s|", token, secondsLeft, strings.Repeat("-", otpPeriod-secondsLeft), strings.Repeat("=", secondsLeft))

	if c.Bool("clip") {
		if err := s.copyToClipboard(ctx, fmt.Sprintf("token for %s", name), []byte(token)); err != nil {
			return s.exitError(ctx, ExitIO, err, "failed to copy to clipboard: %s", err)
		}
	}

	if c.String("qr") != "" {
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
			return s.exitError(ctx, ExitIO, err, "%s", err)
		}

		if err := ioutil.WriteFile(c.String("qr"), qr, 0600); err != nil {
			return s.exitError(ctx, ExitIO, err, "failed to write QR code: %s", err)
		}
	}

	return nil
}
