package action

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/urfave/cli"
)

const (
	totpPeriod = 30 // seconds
)

// TOTP implements time-based OTP token handling
func (s *Action) TOTP(ctx context.Context, c *cli.Context) error {
	name := c.Args().First()
	if name == "" {
		return s.exitError(ctx, ExitUsage, nil, "usage: %s totp [name]", s.Name)
	}

	sec, err := s.Store.Get(ctx, name)
	if err != nil {
		return s.exitError(ctx, ExitDecrypt, err, "failed to get entry '%s': %s", name, err)
	}

	secKey, err := sec.Value("totp")
	if err != nil {
		secKey = sec.Password()
	}

	key, err := otp.NewKeyFromURL(secKey)
	if err != nil {
		return s.exitError(ctx, ExitUnknown, err, "failed get key from URL: %s", err)
	}

	now := time.Now()
	code, err := printCode(key.Secret(), now)
	if err != nil {
		return s.exitError(ctx, ExitIO, err, "failed to encode secret: %s", err)
	}

	_, err = printCode(key.Secret(), now.Add(totpPeriod*time.Second))
	if err != nil {
		return s.exitError(ctx, ExitIO, err, "failed to print encode secret: %s", err)
	}

	if c.Bool("clip") {
		if err := s.copyToClipboard(ctx, fmt.Sprintf("time based token for %s", name), []byte(code)); err != nil {
			return s.exitError(ctx, ExitIO, err, "failed to copy to clipboard: %s", err)
		}
	}

	return nil
}

func printCode(secret string, t time.Time) (string, error) {
	secret = strings.TrimSpace(secret)
	secret = strings.ToUpper(secret)
	code, err := totp.GenerateCodeCustom(secret, t, totp.ValidateOpts{
		Period:    totpPeriod,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		return "", errors.Wrapf(err, "failed to generate OTP code")
	}

	expiresAt := time.Unix(t.Unix()+totpPeriod-(t.Unix()%totpPeriod), 0)
	secondsLeft := int(time.Until(expiresAt).Seconds())

	if secondsLeft <= totpPeriod {
		color.Yellow("%s lasts %ds \t|%s%s|", code, secondsLeft, strings.Repeat("=", totpPeriod-secondsLeft), strings.Repeat("-", secondsLeft))
	} else {
		color.Yellow("%s expires in %ds", code, secondsLeft)
	}
	return code, nil
}
