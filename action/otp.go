package action

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/gokyle/twofactor"
	"github.com/justwatchcom/gopass/store"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/urfave/cli"
)

const (
	// we might want to replace this with the currently un-exported step value
	// from twofactor.FromURL if it gets ever exported
	otpPeriod = 30
)

// OTP implements OTP token handling for TOTP and HOTP
func (s *Action) OTP(ctx context.Context, c *cli.Context) error {
	name := c.Args().First()
	if name == "" {
		return exitError(ctx, ExitUsage, nil, "usage: %s otp [name]", s.Name)
	}
	qrf := c.String("qr")
	clip := c.Bool("clip")

	return s.otp(ctx, name, qrf, clip)
}

func (s *Action) otp(ctx context.Context, name, qrf string, clip bool) error {
	sec, err := s.Store.Get(ctx, name)
	if err != nil {
		return exitError(ctx, ExitDecrypt, err, "failed to get entry '%s': %s", name, err)
	}

	otp, label, err := otpData(ctx, name, sec)
	if err != nil {
		return exitError(ctx, ExitUnknown, err, "No OTP entry found for %s: %s", name, err)
	}
	token := otp.OTP()

	now := time.Now()
	t := now.Add(otpPeriod * time.Second)

	expiresAt := time.Unix(t.Unix()+otpPeriod-(t.Unix()%otpPeriod), 0)
	secondsLeft := int(time.Until(expiresAt).Seconds())

	if secondsLeft >= otpPeriod {
		secondsLeft = secondsLeft - otpPeriod
	}

	out.Yellow(ctx, "%s lasts %ds \t|%s%s|", token, secondsLeft, strings.Repeat("-", otpPeriod-secondsLeft), strings.Repeat("=", secondsLeft))

	if clip {
		if err := copyToClipboard(ctx, fmt.Sprintf("token for %s", name), []byte(token)); err != nil {
			return exitError(ctx, ExitIO, err, "failed to copy to clipboard: %s", err)
		}
		return nil
	}

	if qrf != "" {
		return s.otpWriteQRFile(ctx, otp, label, qrf)
	}
	return nil
}

func otpData(ctx context.Context, name string, sec store.Secret) (twofactor.OTP, string, error) {
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
	secKey, err := sec.Value("totp")
	if err != nil {
		secKey = sec.Password()
	}

	if strings.HasPrefix(secKey, "otpauth://") {
		return twofactor.FromURL(secKey)
	}

	otp, err := twofactor.NewGoogleTOTP(secKey)
	return otp, label, err
}

func (s *Action) otpWriteQRFile(ctx context.Context, otp twofactor.OTP, label, file string) error {
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
		return exitError(ctx, ExitIO, err, "%s", err)
	}

	if err := ioutil.WriteFile(file, qr, 0600); err != nil {
		return exitError(ctx, ExitIO, err, "failed to write QR code: %s", err)
	}
	return nil
}
