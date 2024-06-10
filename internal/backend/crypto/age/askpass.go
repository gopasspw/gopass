package age

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gopasspw/gopass/internal/cache"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/pinentry/cli"
	"github.com/nbutton23/zxcvbn-go"
	"github.com/twpayne/go-pinentry"
	"github.com/zalando/go-keyring"
)

type cacher interface {
	Get(string) (string, bool)
	Set(string, string)
	Remove(string)
	Purge()
}

type osKeyring struct {
	knownKeys map[string]bool
}

func newOsKeyring() *osKeyring {
	return &osKeyring{
		knownKeys: make(map[string]bool),
	}
}

func (o *osKeyring) Get(key string) (string, bool) {
	sec, err := keyring.Get("gopass", key)
	if err != nil {
		debug.Log("failed to get %s from OS keyring: %w", key, err)

		return "", false
	}
	o.knownKeys[name] = true

	return sec, true
}

func (o *osKeyring) Set(name, value string) {
	if err := keyring.Set("gopass", name, value); err != nil {
		debug.Log("failed to set %s: %w", name, err)
	}
	o.knownKeys[name] = true
}

func (o *osKeyring) Remove(name string) {
	if err := keyring.Delete("gopass", name); err != nil {
		debug.Log("failed to remove %s from keyring: %s", name, err)

		return
	}
	o.knownKeys[name] = false
}

func (o *osKeyring) Purge() {
	// purge all known keys. only useful for the REPL case.
	// Does not persist across restarts.
	for k, v := range o.knownKeys {
		if !v {
			continue
		}
		if err := keyring.Delete("gopass", k); err != nil {
			debug.Log("failed to remove %s from keyring: %s", k, err)
		}
	}
}

type askPass struct {
	testing bool
	cache   cacher
}

func newAskPass(ctx context.Context) *askPass {
	a := &askPass{
		cache: cache.NewInMemTTL[string, string](time.Hour, 24*time.Hour),
	}

	if config.Bool(ctx, "age.usekeychain") {
		if err := keyring.Set("gopass", "sentinel", "empty"); err == nil {
			debug.V(1).Log("using OS keychain to cache age credentials")
			a.cache = newOsKeyring()
		}
	}

	return a
}

func (a *askPass) Ping(_ context.Context) error {
	return nil
}

func (a *askPass) Passphrase(key string, reason string, repeat bool) (string, error) {
	if value, found := a.cache.Get(key); found || a.testing {
		debug.V(1).Log("Read value for %s from cache", key)

		return value, nil
	}
	debug.Log("Value for %s not found in cache", key)

	pw, err := a.getPassphrase(reason, repeat)
	if err != nil {
		return "", fmt.Errorf("pinentry error: %w", err)
	}

	debug.V(1).Log("Updated value for %s in cache", key)
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
	} else {
		opts = append(opts,
			pinentry.WithOption(pinentry.OptionAllowExternalPasswordCache),
			pinentry.WithKeyInfo("gopass/age-identities"),
		)
	}

	p, err := pinentry.NewClient(opts...)
	if err != nil {
		debug.Log("Pinentry not found: %q", err)
		// use CLI fallback
		pf := cli.New()
		if repeat {
			_ = pf.Set("REPEAT")
		}

		return pf.GetPIN()
	}
	defer func() {
		_ = p.Close()
	}()

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
