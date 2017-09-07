package root

import (
	"context"

	"github.com/justwatchcom/gopass/config"
	"github.com/pkg/errors"
)

// Config returns this root stores config as a config struct
func (s *Store) Config() *config.Config {
	c := &config.Config{
		Mounts:  make(map[string]string, len(s.mounts)),
		Path:    s.path,
		Version: s.version,
	}
	for alias, sub := range s.mounts {
		c.Mounts[alias] = sub.Path()
	}
	return c
}

// UpdateConfig updates this root-stores internal config and propagates
// those changes to all substores
func (s *Store) UpdateConfig(ctx context.Context, cfg *config.Config) error {
	if cfg == nil {
		return errors.Errorf("invalid config")
	}
	s.path = cfg.Path

	// add any missing mounts
	for alias, path := range cfg.Mounts {
		if _, found := s.mounts[alias]; !found {
			if err := s.addMount(ctx, alias, path); err != nil {
				return errors.Wrapf(err, "failed to add mount '%s' to '%s'", alias, path)
			}
		}
	}

	// propagate any config changes to our substores
	if err := s.store.UpdateConfig(cfg); err != nil {
		return errors.Wrapf(err, "failed to update config for root store")
	}
	for _, sub := range s.mounts {
		if err := sub.UpdateConfig(cfg); err != nil {
			return errors.Wrapf(err, "failed to update config for sub store %s", sub.Alias)
		}
	}

	return nil
}

// Path returns the store path
func (s *Store) Path() string {
	return s.path
}

// Alias always returns an empty string
func (s *Store) Alias() string {
	return ""
}
