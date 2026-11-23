package age

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/gopasspw/gopass/internal/cache"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/pinentry/cli"
	"github.com/gopasspw/gopass/pkg/termio"
	"github.com/twpayne/go-pinentry/v4"
	"github.com/zalando/go-keyring"
)

type cacher interface {
	Get(string) (string, bool)
	Set(context.Context, string, string)
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
		debug.Log("failed to get %s from OS keyring: %v", key, err)

		return "", false
	}
	o.knownKeys[key] = true

	return sec, true
}

func (o *osKeyring) Set(ctx context.Context, name, value string) {
	if err := keyring.Set("gopass", name, value); err != nil {
		debug.Log("failed to set %s: %v", name, err)
		out.Warningf(ctx, "Failed to cache passphrase in OS keyring: %s", err)

		return
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

var keyringWarningOnce sync.Once

func newAskPass(ctx context.Context) *askPass {
	a := &askPass{
		cache: cache.NewInMemTTL[string, string](time.Hour, 24*time.Hour),
	}

	if config.Bool(ctx, "age.usekeychain") {
		if err := keyring.Set("gopass", "sentinel", "empty"); err == nil {
			debug.V(1).Log("using OS keychain to cache age credentials")
			a.cache = newOsKeyring()
		} else {
			keyringWarningOnce.Do(func() {
				out.Warningf(ctx, "OS keyring is not available. Passphrase caching will not persist. Disable age.usekeychain or install a keyring provider (e.g. gnome-keyring): %s", err)
			})
		}
	}

	return a
}

func (a *askPass) Ping(_ context.Context) error {
	return nil
}

func (a *askPass) Passphrase(ctx context.Context, key string, reason string, repeat bool) (string, error) {
	if value, found := a.cache.Get(key); found || a.testing {
		debug.V(1).Log("Read value for %s from cache", key)

		return value, nil
	}
	debug.V(1).Log("Value for %s not found in cache", key)

	pw, err := a.getPassphrase(reason, repeat)
	if err != nil {
		return "", fmt.Errorf("pinentry error: %w", err)
	}

	debug.V(1).Log("Updated value for %s in cache", key)
	a.cache.Set(ctx, key, pw)

	return pw, nil
}

func (a *askPass) getPassphrase(reason string, repeat bool) (string, error) {
	if os.Getenv("GOPASS_AGE_STDIN_PASSPHRASE") != "" {
		debug.Log("GOPASS_AGE_STDIN_PASSPHRASE is set, using CLI fallback")
		pf := cli.New()
		if repeat {
			_ = pf.Set("REPEAT")
		}

		ctx := ctxutil.WithTerminal(context.Background(), false)
		pw, err := termio.AskForPassword(ctx, reason, repeat)
		if err != nil {
			return "", fmt.Errorf("failed to ask for passphrase: %w", err)
		}

		return pw, nil
	}

	opts := []pinentry.ClientOption{
		pinentry.WithBinaryNameFromGnuPGAgentConf(),
		pinentry.WithDesc(strings.TrimSuffix(reason, ":") + "."),
		pinentry.WithPrompt("Passphrase:"),
		pinentry.WithTitle("gopass"),
	}
	opts = append(opts, pinentryTTYOptions()...)
	if repeat {
		opts = append(opts, pinentry.WithRepeat("Confirm"))
	} else {
		opts = append(
			opts,
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

	result, err := p.GetPIN()
	if err != nil {
		return "", fmt.Errorf("pinentry error: %w", err)
	}

	return result.PIN, nil
}

func (a *askPass) Remove(key string) {
	a.cache.Remove(key)
}

func (a *askPass) Lock() {
	a.cache.Purge()
}

// pinentryTTYOptions returns pinentry client options for configuring the TTY
// (ttyname, ttytype, and lc-ctype) on Unix systems. On Windows it returns nil.
//
// Unlike pinentry.WithGPGTTY() (which only checks the GPG_TTY environment
// variable and sends nothing when it is unset, causing pinentry-curses and
// pinentry-tty to fail with ERR 83918950 Inappropriate ioctl for device on the
// Assuan stdin pipe), this falls back to /dev/tty whenever the process has a
// controlling terminal.
func pinentryTTYOptions() []pinentry.ClientOption {
	if runtime.GOOS == "windows" {
		return nil
	}

	return pinentryTTYOptionsFor(os.Getenv, isCharDevice, hasControllingTTY)
}

func pinentryTTYOptionsFor(getenv func(string) string, isTTY func(string) bool, hasDevTTY func() bool) []pinentry.ClientOption {
	var opts []pinentry.ClientOption

	ttyName := strings.TrimSpace(getenv("GPG_TTY"))
	if ttyName == "" || !isTTY(ttyName) {
		if hasDevTTY() {
			ttyName = "/dev/tty"
		} else {
			ttyName = ""
		}
	}
	if ttyName != "" {
		opts = append(opts, pinentry.WithOption(pinentry.OptionTTYName+"="+ttyName))
	}

	if term := strings.TrimSpace(getenv("TERM")); term != "" {
		opts = append(opts, pinentry.WithOption(pinentry.OptionTTYType+"="+term))
	}

	for _, envKey := range []string{"LC_ALL", "LC_CTYPE", "LANG"} {
		if v := strings.TrimSpace(getenv(envKey)); v != "" {
			opts = append(opts, pinentry.WithOption(pinentry.OptionLCCType+"="+v))

			break
		}
	}

	return opts
}

func isCharDevice(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}

	return fi.Mode()&os.ModeCharDevice != 0
}

func hasControllingTTY() bool {
	f, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return false
	}
	_ = f.Close()

	return true
}
