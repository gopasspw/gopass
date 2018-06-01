package sub

import (
	"context"
	"fmt"

	"github.com/gopasspw/gopass/pkg/backend"
	"github.com/gopasspw/gopass/pkg/backend/storage/fs"
	kvconsul "github.com/gopasspw/gopass/pkg/backend/storage/kv/consul"
	"github.com/gopasspw/gopass/pkg/backend/storage/kv/inmem"
	"github.com/gopasspw/gopass/pkg/config/secrets"
	"github.com/gopasspw/gopass/pkg/out"

	"github.com/pkg/errors"
)

func (s *Store) initStorageBackend(ctx context.Context) error {
	switch s.url.Storage {
	case backend.FS:
		s.storage = fs.New(s.url.Path)
		out.Debug(ctx, "Using Storage Backend: %s", s.storage.String())
	case backend.InMem:
		out.Debug(ctx, "Using Storage Backend: inmem")
		s.storage = inmem.New()
	case backend.Consul:
		out.Debug(ctx, "Using Storage Backend: consul")
		token := s.url.Query.Get("token")
		key := fmt.Sprintf("consul-%s-%s", s.alias, s.url.String())
		if token == "" {
			out.Debug(ctx, "Requesting token from secrets config: %s", key)
			t, err := s.loadSecret(ctx, key)
			if err != nil {
				return errors.Wrapf(err, "failed to load token from secrets config: %s", err)
			}
			out.Debug(ctx, "Got token from secrets config: '%s'", t)
			token = t
		}
		out.Debug(ctx, "Consul-Token: '%s'", token)
		store, err := kvconsul.New(s.url.Host+":"+s.url.Port, s.url.Path, s.url.Query.Get("datacenter"), token)
		if err != nil {
			_ = s.agent.Remove(ctx, key)
			return err
		}
		// test connection and save token if it works
		if err := store.Available(ctx); err != nil {
			out.Debug(ctx, "Consul access not working. removing saved token")
			_ = s.eraseSecret(ctx, key)
			return err
		}
		out.Debug(ctx, "Consul access OK. saving token")
		if err := s.storeSecret(ctx, key, token); err != nil {
			return err
		}
		s.storage = store
	default:
		return fmt.Errorf("unknown storage backend")
	}
	return nil
}

func (s *Store) storeSecret(ctx context.Context, key, value string) error {
	pw, err := s.agent.Passphrase(ctx, "config.sec", "Please enter passphrase to (un)lock config secrets")
	if err != nil {
		return err
	}
	seccfg, err := secrets.New(s.cfgdir, pw)
	if err != nil {
		return err
	}
	return seccfg.Set(key, value)
}

func (s *Store) eraseSecret(ctx context.Context, key string) error {
	pw, err := s.agent.Passphrase(ctx, "config.sec", "Please enter passphrase to (un)lock config secrets")
	if err != nil {
		return err
	}
	seccfg, err := secrets.New(s.cfgdir, pw)
	if err != nil {
		return err
	}
	_ = s.agent.Remove(ctx, key)
	return seccfg.Unset(key)
}

func (s *Store) loadSecret(ctx context.Context, key string) (string, error) {
	pw, err := s.agent.Passphrase(ctx, "config.sec", "Please enter passphrase to (un)lock config secrets")
	if err != nil {
		return "", err
	}
	seccfg, err := secrets.New(s.cfgdir, pw)
	if err != nil {
		return "", err
	}
	t, err := seccfg.Get(key)
	if err == nil && t != "" {
		return t, nil
	}
	t, err = s.agent.Passphrase(ctx, key, "Please enter the secret "+key)
	if err != nil {
		return "", err
	}
	_ = s.agent.Remove(ctx, key)
	err = seccfg.Set(key, t)
	return t, err
}
