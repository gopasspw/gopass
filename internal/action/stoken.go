package action

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gopasspw/gopass/internal/clipboard"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/stoken"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/urfave/cli/v2"
)

// SToken generates RSA tokens with libstoken
func (s *Action) SToken(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	name := c.Args().First()
	if name == "" {
		return ExitError(ExitUsage, nil, "Usage: %s stoken <NAME>", s.Name)
	}

	clip := c.Bool("clip")
	unsafe := c.Bool("unsafe")
	pw := c.Bool("password")
	pin := c.String("pin")
	devid := c.String("devid")
	seedpw := c.String("seedpw")

	return s.stoken(ctx, name, pin, devid, seedpw, clip, unsafe, pw, true)
}

func (s *Action) stoken(ctx context.Context, name, pin, devid, seedpw string, clip, unsafe, pw, recurse bool) error {
	sec, err := s.Store.Get(ctx, name)
	if err != nil {
		return s.stokenError(ctx, name, pin, devid, seedpw, clip, unsafe, pw, recurse, err)
	}
	if unsafe {
		out.Print(ctx, "%s", sec.Get("Password"))
		return nil
	}
	if pin == "" {
		pin = sec.Get("Pin")
	}
	if devid == "" {
		devid = sec.Get("DeviceID")
	}
	if seedpw == "" {
		seedpw = sec.Get("SeedPassword")
	}

	now := time.Now()
	token, interval, err := stoken.Compute(now, pin, devid, seedpw, sec)
	if err != nil {
		return ExitError(ExitUnknown, err, "No OTP entry found for %s: %s", name, err)
	}
	timeRemaining := interval - (now.Unix() % interval)
	expTime := now.Add(time.Second * time.Duration(timeRemaining))
	secondsLeft := int64(time.Until(expTime).Seconds())

	if pw {
		out.Print(ctx, "%s", token)
	} else {
		out.Yellow(ctx, "%s lasts %ds \t|%s%s|", token, secondsLeft, strings.Repeat("-", int(interval-secondsLeft)), strings.Repeat("=", int(secondsLeft)))
	}

	if clip {
		if err := clipboard.CopyTo(ctx, fmt.Sprintf("token for %s", name), []byte(token)); err != nil {
			return ExitError(ExitIO, err, "failed to copy to clipboard: %s", err)
		}
		return nil
	}
	return nil
}

func (s *Action) stokenError(ctx context.Context, name, pin, devid, seedpw string, clip, unsafe, pw, recurse bool, err error) error {
	if err != store.ErrNotFound || !recurse || !ctxutil.IsTerminal(ctx) {
		return ExitError(ExitUnknown, err, "failed to retrieve secret '%s': %s", name, err)
	}
	out.Yellow(ctx, "Entry '%s' not found. Starting search...", name)
	cb := func(ctx context.Context, c *cli.Context, name string, recurse bool) error {
		return s.stoken(ctx, name, pin, devid, seedpw, clip, unsafe, pw, false)
	}
	if err := s.find(ctxutil.WithFuzzySearch(ctx, false), nil, name, cb); err != nil {
		return ExitError(ExitNotFound, err, "%s", err)
	}
	return nil
}
