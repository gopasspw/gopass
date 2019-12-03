package consul

import (
	"context"
	"fmt"

	"github.com/gopasspw/gopass/pkg/agent/client"
	"github.com/gopasspw/gopass/pkg/backend"
	"github.com/gopasspw/gopass/pkg/config/secrets"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/pkg/errors"
)

const (
	name = "consul"
)

func init() {
	backend.RegisterStorage(backend.Consul, name, &loader{})
}

type loader struct{}

// New implements backend.StorageLoader
func (l loader) New(ctx context.Context, url *backend.URL) (backend.Storage, error) {
	alias := ctxutil.GetAlias(ctx)
	agent := client.GetClient(ctx)
	cfgdir := ctxutil.GetConfigDir(ctx)

	out.Debug(ctx, "Using Storage Backend: %s", name)
	token := url.Query.Get("token")
	key := fmt.Sprintf("consul-%s-%s", alias, url.String())
	if token == "" {
		out.Debug(ctx, "Requesting token from secrets config: %s", key)
		t, err := l.loadSecret(ctx, key, cfgdir, agent)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to load token from secrets config: %s", err)
		}
		out.Debug(ctx, "Got token from secrets config: '%s'", t)
		token = t
	}
	out.Debug(ctx, "Consul-Token: '%s'", token)
	store, err := New(url.Host+":"+url.Port, url.Path, url.Query.Get("datacenter"), token)
	if err != nil {
		_ = agent.Remove(ctx, key)
		return nil, err
	}
	// test connection and save token if it works
	if err := store.Available(ctx); err != nil {
		out.Debug(ctx, "Consul access not working. removing saved token")
		_ = l.eraseSecret(ctx, key, cfgdir, agent)
		return nil, err
	}
	out.Debug(ctx, "Consul access OK. saving token")
	if err := l.storeSecret(ctx, key, token, cfgdir, agent); err != nil {
		return nil, err
	}
	return store, nil
}

func (l loader) storeSecret(ctx context.Context, key, value, cfgdir string, agent *client.Client) error {
	pw, err := agent.Passphrase(ctx, "config.sec", "Please enter passphrase to (un)lock config secrets")
	if err != nil {
		return err
	}
	seccfg, err := secrets.New(cfgdir, pw)
	if err != nil {
		return err
	}
	return seccfg.Set(key, value)
}

func (l loader) eraseSecret(ctx context.Context, key, cfgdir string, agent *client.Client) error {
	pw, err := agent.Passphrase(ctx, "config.sec", "Please enter passphrase to (un)lock config secrets")
	if err != nil {
		return err
	}
	seccfg, err := secrets.New(cfgdir, pw)
	if err != nil {
		return err
	}
	_ = agent.Remove(ctx, key)
	return seccfg.Unset(key)
}

func (l loader) loadSecret(ctx context.Context, key, cfgdir string, agent *client.Client) (string, error) {
	pw, err := agent.Passphrase(ctx, "config.sec", "Please enter passphrase to (un)lock config secrets")
	if err != nil {
		return "", err
	}
	seccfg, err := secrets.New(cfgdir, pw)
	if err != nil {
		return "", err
	}
	t, err := seccfg.Get(key)
	if err == nil && t != "" {
		return t, nil
	}
	t, err = agent.Passphrase(ctx, key, "Please enter the secret "+key)
	if err != nil {
		return "", err
	}
	_ = agent.Remove(ctx, key)
	err = seccfg.Set(key, t)
	return t, err
}

func (l loader) String() string {
	return name
}
