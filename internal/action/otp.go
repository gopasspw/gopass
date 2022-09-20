package action

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gopasspw/gopass/internal/action/exit"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/pkg/clipboard"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/otp"
	"github.com/gopasspw/gopass/pkg/termio"
	"github.com/mattn/go-tty"
	"github.com/pquerna/otp/hotp"
	"github.com/pquerna/otp/totp"
	"github.com/urfave/cli/v2"
)

// OTP implements OTP token handling for TOTP and HOTP.
func (s *Action) OTP(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	name := c.Args().First()
	if name == "" {
		return exit.Error(exit.Usage, nil, "Usage: %s otp <NAME>", s.Name)
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
			return // returning not to leak the goroutine.
		default:
			// we don't want to block if not cancelled.
		}
		if tt.After(expiresAt) {
			return
		}
		bar.Inc()
	}
}

func waitForKeyPress(ctx context.Context, cancel context.CancelFunc) (func(), func()) {
	tty1, err := tty.Open()
	if err != nil {
		out.Errorf(ctx, "Unexpected error opening tty: %v", err)
		cancel()
	}

	return func() {
			for {
				select {
				case <-ctx.Done():
					return // returning not to leak the goroutine.
				default:
				}

				r, err := tty1.ReadRune()
				if err != nil {
					out.Errorf(ctx, "Unexpected error opening tty: %v", err)
				}

				if r == 'q' || r == 'x' || err != nil {
					cancel()

					return
				}
			}
		}, func() {
			_ = tty1.Close()
		}
}

// nolint: cyclop
func (s *Action) otp(ctx context.Context, name, qrf string, clip, pw, recurse bool) error {
	sec, err := s.Store.Get(ctx, name)
	if err != nil {
		return s.otpHandleError(ctx, name, qrf, clip, pw, recurse, err)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	skip := ctxutil.IsHidden(ctx) || pw || qrf != "" || !ctxutil.IsTerminal(ctx) || !ctxutil.IsInteractive(ctx) || clip
	if !skip {
		// let us monitor key presses for cancellation:.
		runFn, cleanupFn := waitForKeyPress(ctx, cancel)
		go runFn()
		defer cleanupFn()
	}

	// only used for the HOTP case as a fallback
	var counter uint64 = 1
	if sv, found := sec.Get("counter"); found && sv != "" {
		if iv, err := strconv.ParseUint(sv, 10, 64); iv != 0 && err == nil {
			counter = iv
		}
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		two, err := otp.Calculate(name, sec)
		if err != nil {
			return exit.Error(exit.Unknown, err, "No OTP entry found for %s: %s", name, err)
		}

		var token string
		switch two.Type() {
		case "totp":
			token, err = totp.GenerateCodeCustom(two.Secret(), time.Now(), totp.ValidateOpts{
				Period:    uint(two.Period()),
				Skew:      1,
				Digits:    two.Digits(),
				Algorithm: two.Algorithm(),
			})
			if err != nil {
				return exit.Error(exit.Unknown, err, "Failed to compute OTP token for %s: %s", name, err)
			}
		case "hotp":
			token, err = hotp.GenerateCodeCustom(two.Secret(), counter, hotp.ValidateOpts{
				Digits:    two.Digits(),
				Algorithm: two.Algorithm(),
			})
			if err != nil {
				return exit.Error(exit.Unknown, err, "Failed to compute OTP token for %s: %s", name, err)
			}
			counter++
			_ = sec.Set("counter", strconv.Itoa(int(counter)))
			if err := s.Store.Set(ctx, name, sec); err != nil {
				out.Errorf(ctx, "Failed to persist counter value: %s", err)
			}
			debug.Log("Saved counter as %d", counter)
		}

		now := time.Now()
		expiresAt := now.Add(time.Duration(two.Period()) * time.Second).Truncate(time.Duration(two.Period()) * time.Second)
		secondsLeft := int(time.Until(expiresAt).Seconds())
		bar := termio.NewProgressBar(int64(secondsLeft))
		bar.Hidden = skip

		debug.Log("OTP period: %ds", two.Period())

		if clip {
			if err := clipboard.CopyTo(ctx, fmt.Sprintf("token for %s", name), []byte(token), s.cfg.ClipTimeout); err != nil {
				return exit.Error(exit.IO, err, "failed to copy to clipboard: %s", err)
			}

			return nil
		}

		// check if we are in "password only" or in "qr code" mode or being redirected to a pipe.
		if pw || qrf != "" || !ctxutil.IsTerminal(ctx) {
			out.Printf(ctx, "%s", token)
			cancel()
		} else { // if not then we want to print a progress bar with the expiry time.
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
			return otp.WriteQRFile(two, qrf)
		}

		// let us wait until next OTP code:.
		for {
			select {
			case <-ctx.Done():
				bar.Done()
				cancel()

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
	if !errors.Is(err, store.ErrNotFound) || !recurse || !ctxutil.IsTerminal(ctx) {
		return exit.Error(exit.Unknown, err, "failed to retrieve secret %q: %s", name, err)
	}

	out.Printf(ctx, "Entry %q not found. Starting search...", name)
	cb := func(ctx context.Context, c *cli.Context, name string, recurse bool) error {
		return s.otp(ctx, name, qrf, clip, pw, false)
	}
	if err := s.find(ctx, nil, name, cb, false); err != nil {
		return exit.Error(exit.NotFound, err, "%s", err)
	}

	return nil
}
