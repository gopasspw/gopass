package age

import (
	"context"
	"fmt"
	"time"

	"github.com/gopasspw/gopass/internal/cache"
	"github.com/gopasspw/gopass/internal/debug"
	"github.com/gopasspw/gopass/pkg/pinentry"
)

type piner interface {
	Close()
	Confirm() bool
	Set(string, string) error
	GetPin() ([]byte, error)
}

type cacher interface {
	Get(string) (string, bool)
	Set(string, string)
	Remove(string)
}

type askPass struct {
	testing  bool
	cache    cacher
	pinentry func() (piner, error)
}

var (
	defaultAskPass = newAskPass()
)

func newAskPass() *askPass {
	return &askPass{
		cache: cache.NewInMemTTL(time.Hour, 24*time.Hour),
		pinentry: func() (piner, error) {
			return pinentry.New()
		},
	}
}

func (a *askPass) Ping(_ context.Context) error {
	return nil
}

func (a *askPass) Passphrase(key string, reason string) (string, error) {
	if value, found := a.cache.Get(key); found || a.testing {
		debug.Log("Read value for %s from cache", key)
		return value, nil
	}
	debug.Log("Value for %s not found in cache", key)

	pi, err := a.pinentry()
	if err != nil {
		return "", fmt.Errorf("pinentry Error: %s", err)
	}
	defer pi.Close()

	_ = pi.Set("title", "gopass")
	_ = pi.Set("desc", "Need your passphrase "+reason)
	_ = pi.Set("prompt", "Please enter your passphrase:")
	_ = pi.Set("ok", "OK")

	pw, err := pi.GetPin()
	if err != nil {
		return "", fmt.Errorf("pinentry Error: %s", err)
	}

	pass := string(pw)
	debug.Log("Updated value for %s in cache", key)
	a.cache.Set(key, pass)
	return pass, nil
}

func (a *askPass) Remove(key string) {
	a.cache.Remove(key)
}
