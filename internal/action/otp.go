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

	"github.com/mattn/go-tty"
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

func tickingBar(ctx context.Context, expiresAt time.Time, bar *termio.ProgressBar) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for tt := range ticker.C {
		select {
		case <-ctx.Done():
			return // returning not to leak the goroutine
		default:
			// we don't want to block if not cancelled
		}
		if tt.After(expiresAt) {
			return
		}
		bar.Inc()
	}
}

func waitForKeyPress(ctx context.Context, cancel context.CancelFunc) {
	tty, err := tty.Open()
	if err != nil {
		out.Errorf(ctx, "Unexpected error opening tty: %v", err)
		cancel()
	}
	defer tty.Close()
	for {
		select {
		case <-ctx.Done():
			return // returning not to leak the goroutine
		default:
		}
		r, err := tty.ReadRune()
		if err != nil {
			out.Errorf(ctx, "Unexpected error opening tty: %v", err)
		}
		if r == 'q' || r == 'x' || err != nil {
			cancel()
			return
		}
	}
}

func (s *Action) otp(ctx context.Context, name, qrf string, clip, pw, recurse bool) error {
	sec, err := s.Store.Get(ctx, name)
	if err != nil {
		return s.otpHandleError(ctx, name, qrf, clip, pw, recurse, err)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	skip := ctxutil.IsHidden(ctx) || pw || qrf != "" || out.OutputIsRedirected() || !ctxutil.IsInteractive(ctx)
	if !skip {
		// let us monitor key presses for cancellation:
		go waitForKeyPress(ctx, cancel)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		two, label, err := otp.Calculate(name, sec)
		if err != nil {
			return ExitError(ExitUnknown, err, "No OTP entry found for %s: %s", name, err)
		}
		token := two.OTP()

		now := time.Now()
		expiresAt := now.Add(otpPeriod * time.Second).Truncate(otpPeriod * time.Second)
		secondsLeft := int(time.Until(expiresAt).Seconds())
		bar := termio.NewProgressBar(int64(secondsLeft))
		bar.Hidden = skip

		if clip {
			if err := clipboard.CopyTo(ctx, fmt.Sprintf("token for %s", name), []byte(token), s.cfg.ClipTimeout); err != nil {
				return ExitError(ExitIO, err, "failed to copy to clipboard: %s", err)
			}
		}

		// check if we are in "password only" or in "qr code" mode or being redirected to a pipe
		if pw || qrf != "" || out.OutputIsRedirected() {
			out.Printf(ctx, "%s", token)
			cancel()
		} else { // if not then we want to print a progress bar with the expiry time
			out.Printf(ctx, "%s", token)
			out.Warningf(ctx, "([q] to stop. -o flag to avoid.) This OTP password still lasts for:", nil)

			if bar.Hidden {
				cancel()
			} else {
				bar.Set(0)
				go tickingBar(ctx, expiresAt, bar)
			}
		}

		if qrf != "" {
			return otp.WriteQRFile(two, label, qrf)
		}

		// let us wait until next OTP code:
		for {
			select {
			case <-ctx.Done():
				bar.Done()
				return nil
			default:
				time.Sleep(time.Millisecond * 500)
			}
			if time.Now().After(expiresAt) {
				bar.Done()
				break
			}
		}
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
