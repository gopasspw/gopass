package action

import (
	"errors"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/urfave/cli"
)

const (
	totpPeriod = 30 // seconds
)

// TOTP implements time-based OTP token handling
func (s *Action) TOTP(c *cli.Context) error {
	name := c.Args().First()
	if name == "" {
		return errors.New("provide a password name")
	}

	content, err := s.Store.Get(name)
	if err != nil {
		return err
	}

	key, err := otp.NewKeyFromURL(string(content))
	if err != nil {
		return err
	}

	now := time.Now()
	_, err = printCode(key.Secret(), now)
	if err != nil {
		return err
	}

	_, err = printCode(key.Secret(), now.Add(totpPeriod*time.Second))
	if err != nil {
		return err
	}

	return nil
}

func printCode(secret string, t time.Time) (int, error) {
	secret = strings.TrimSpace(secret)
	secret = strings.ToUpper(secret)
	code, err := totp.GenerateCodeCustom(secret, t, totp.ValidateOpts{
		Period:    totpPeriod,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		return 0, err
	}

	expiresAt := time.Unix(t.Unix()+totpPeriod-(t.Unix()%totpPeriod), 0)
	secondsLeft := int(expiresAt.Sub(time.Now()) / time.Second)

	if secondsLeft <= totpPeriod {
		color.Yellow("%s lasts %ds \t|%s%s|", code, secondsLeft, strings.Repeat("=", totpPeriod-secondsLeft), strings.Repeat("-", secondsLeft))
	} else {
		color.Yellow("%s expires in %ds", code, secondsLeft)
	}
	return secondsLeft, nil
}
