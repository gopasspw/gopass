package action

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gopasspw/gopass/pkg/clipboard"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/otp"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/pkg/store"

	"gopkg.in/urfave/cli.v1"
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
	pw := c.Bool("password")

	return s.otp(ctx, c, name, qrf, clip, pw, true)
}

func (s *Action) otp(ctx context.Context, c *cli.Context, name, qrf string, clip, pw, recurse bool) error {
	sec, ctx, err := s.Store.GetContext(ctx, name)
	if err != nil {
		return s.otpHandleError(ctx, c, name, qrf, clip, pw, recurse, err)
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
		secondsLeft -= otpPeriod
	}

	if pw {
		out.Print(ctx, "%s", token)
	} else {
		out.Yellow(ctx, "%s lasts %ds \t|%s%s|", token, secondsLeft, strings.Repeat("-", otpPeriod-secondsLeft), strings.Repeat("=", secondsLeft))
	}

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

func (s *Action) otpHandleError(ctx context.Context, c *cli.Context, name, qrf string, clip, pw, recurse bool, err error) error {
	if err != store.ErrNotFound || !recurse || !ctxutil.IsTerminal(ctx) {
		return ExitError(ctx, ExitUnknown, err, "failed to retrieve secret '%s': %s", name, err)
	}
	out.Yellow(ctx, "Entry '%s' not found. Starting search...", name)
	cb := func(ctx context.Context, c *cli.Context, name string, recurse bool) error {
		return s.otp(ctx, c, name, qrf, clip, pw, false)
	}
	if err := s.find(ctxutil.WithFuzzySearch(ctx, false), nil, name, cb); err == nil {
		return nil
	}
	os.Exit(ExitNotFound)
	return nil
}
