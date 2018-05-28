package action

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gopasspw/gopass/pkg/clipboard"
	"github.com/gopasspw/gopass/pkg/otp"
	"github.com/gopasspw/gopass/pkg/out"

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
		return ExitError(ctx, ExitUsage, nil, "Usage: %s otp <NAME>", s.Name)
	}

	qrf := c.String("qr")
	clip := c.Bool("clip")

	return s.otp(ctx, name, qrf, clip)
}

func (s *Action) otp(ctx context.Context, name, qrf string, clip bool) error {
	sec, err := s.Store.Get(ctx, name)
	if err != nil {
		return ExitError(ctx, ExitDecrypt, err, "failed to get entry '%s': %s", name, err)
	}

	two, label, err := otp.Calculate(ctx, name, sec)
	if err != nil {
		return ExitError(ctx, ExitUnknown, err, "No OTP entry found for %s: %s", name, err)
	}
	token := two.OTP()

	now := time.Now()
	t := now.Add(otpPeriod * time.Second)

	expiresAt := time.Unix(t.Unix()+otpPeriod-(t.Unix()%otpPeriod), 0)
	secondsLeft := int(time.Until(expiresAt).Seconds())

	if secondsLeft >= otpPeriod {
		secondsLeft = secondsLeft - otpPeriod
	}

	out.Yellow(ctx, "%s lasts %ds \t|%s%s|", token, secondsLeft, strings.Repeat("-", otpPeriod-secondsLeft), strings.Repeat("=", secondsLeft))

	if clip {
		if err := clipboard.CopyTo(ctx, fmt.Sprintf("token for %s", name), []byte(token)); err != nil {
			return ExitError(ctx, ExitIO, err, "failed to copy to clipboard: %s", err)
		}
		return nil
	}

	if qrf != "" {
		return otp.WriteQRFile(ctx, two, label, qrf)
	}
	return nil
}
