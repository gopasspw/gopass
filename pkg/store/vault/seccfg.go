package vault

import (
	"context"

	"github.com/justwatchcom/gopass/pkg/config/secrets"
)

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
