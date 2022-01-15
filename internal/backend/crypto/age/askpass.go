package age

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gopasspw/gopass/internal/cache"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/pinentry/cli"
	"github.com/nbutton23/zxcvbn-go"
	"github.com/twpayne/go-pinentry"
)

type cacher interface {
	Get(string) (string, bool)
	Set(string, string)
	Remove(string)
	Purge()
}

type askPass struct {
	testing bool
	cache   cacher
}

var (
	// DefaultAskPass is the default password cache.
	DefaultAskPass = newAskPass()
)

func newAskPass() *askPass {
	return &askPass{
		cache: cache.NewInMemTTL[string, string](time.Hour, 24*time.Hour),
	}
}

func (a *askPass) Ping(_ context.Context) error {
	return nil
}

func (a *askPass) Passphrase(key string, reason string, repeat bool) (string, error) {
	if value, found := a.cache.Get(key); found || a.testing {
		debug.Log("Read value for %s from cache", key)
		return value, nil
	}
	debug.Log("Value for %s not found in cache", key)

	pw, err := a.getPassphrase(reason, repeat)
	if err != nil {
		return "", fmt.Errorf("pinentry error: %w", err)
	}

	debug.Log("Updated value for %s in cache", key)
	a.cache.Set(key, pw)
	return pw, nil
}

func (a *askPass) getPassphrase(reason string, repeat bool) (string, error) {
	opts := []pinentry.ClientOption{
		pinentry.WithBinaryNameFromGnuPGAgentConf(),
		pinentry.WithDesc(strings.TrimSuffix(reason, ":") + "."),
		pinentry.WithGPGTTY(),
		pinentry.WithPrompt("Passphrase:"),
		pinentry.WithTitle("gopass"),
	}
	if repeat {
		opts = append(opts, pinentry.WithOption("REPEAT=Confirm"))
		opts = append(opts, pinentry.WithQualityBar(func(s string) (int, bool) {
			match := zxcvbn.PasswordStrength(s, nil)
			return match.Score, true
		}))
	}

	p, err := pinentry.NewClient(opts...)
	if err != nil {
		debug.Log("Pinentry not found: %q", err)
		// use CLI fallback
		pf := cli.New()
		if repeat {
			pf.Set("REPEAT")
		}
		return pf.GetPIN()
	}
	defer p.Close()

	pw, _, err := p.GetPIN()
	if err != nil {
		return "", fmt.Errorf("pinentry error: %w", err)
	}

	return pw, nil
}

func (a *askPass) Remove(key string) {
	a.cache.Remove(key)
}

// Lock flushes the password cache.
func (a *Age) Lock() {
	a.askPass.cache.Purge()
}
