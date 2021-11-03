package action

import (
	"context"
	"fmt"
	"time"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/pkg/clipboard"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/otp"
	"github.com/gopasspw/gopass/pkg/termio"

	"github.com/urfave/cli/v2"
)

const (
	// we might want to replace this with the currently un-exported step value
	// from twofactor.FromURL if it gets ever exported
	otpPeriod = 30
)

// OTP implements OTP token handling for TOTP and HOTP
func (s *Action) OTP(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	name := c.Args().First()
	if name == "" {
		return ExitError(ExitUsage, nil, "Usage: %s otp <NAME>", s.Name)
	}

	qrf := c.String("qr")
	clip := c.Bool("clip")
	pw := c.Bool("password")

	return s.otp(ctx, name, qrf, clip, pw, true)
}

func (s *Action) otp(ctx context.Context, name, qrf string, clip, pw, recurse bool) error {
	sec, err := s.Store.Get(ctx, name)
	if err != nil {
		return s.otpHandleError(ctx, name, qrf, clip, pw, recurse, err)
	}
	two, label, err := otp.Calculate(name, sec)
	if err != nil {
		return ExitError(ExitUnknown, err, "No OTP entry found for %s: %s", name, err)
	}
	token := two.OTP()

	now := time.Now()
	expiresAt := now.Add(otpPeriod * time.Second).Truncate(otpPeriod * time.Second)
	secondsLeft := int(time.Until(expiresAt).Seconds())

	if clip {
		if err := clipboard.CopyTo(ctx, fmt.Sprintf("token for %s", name), []byte(token), s.cfg.ClipTimeout); err != nil {
			return ExitError(ExitIO, err, "failed to copy to clipboard: %s", err)
		}
	}

	done := make(chan bool)
	skip := false
	// check if we are in "password only" or in "qr code" mode or being redirected to a pipe
	if pw || qrf != "" || out.OutputIsRedirected() {
		out.Printf(ctx, "%s", token)
		skip = true
	} else { // if not then we want to print a progress bar with the expiry time
		out.Printf(ctx, "%s", token)
		out.Warningf(ctx, "This OTP password still lasts for:", nil)
		bar := termio.NewProgressBar(int64(secondsLeft))
		bar.Hidden = ctxutil.IsHidden(ctx)
		if bar.Hidden {
			skip = true
		} else {
			bar.Set(0)
			go func() {
				ticker := time.NewTicker(1 * time.Second)
				defer ticker.Stop()
				for tt := range ticker.C {
					if tt.After(expiresAt) {
						bar.Done()
						done <- true
						return
					}
					bar.Inc()
				}
			}()
		}
	}

	if qrf != "" {
		return otp.WriteQRFile(two, label, qrf)
	}

	// we need to return if we are skipping, to avoid a deadlock in select
	if skip {
		return nil
	}

	// we wait until our ticker is done or we get a cancelation
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return termio.ErrAborted
	}
}

func (s *Action) otpHandleError(ctx context.Context, name, qrf string, clip, pw, recurse bool, err error) error {
	if err != store.ErrNotFound || !recurse || !ctxutil.IsTerminal(ctx) {
		return ExitError(ExitUnknown, err, "failed to retrieve secret %q: %s", name, err)
	}
	out.Printf(ctx, "Entry %q not found. Starting search...", name)
	cb := func(ctx context.Context, c *cli.Context, name string, recurse bool) error {
		return s.otp(ctx, name, qrf, clip, pw, false)
	}
	if err := s.find(ctx, nil, name, cb, false); err != nil {
		return ExitError(ExitNotFound, err, "%s", err)
	}
	return nil
}
