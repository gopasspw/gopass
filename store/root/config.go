package root

import (
	"context"

	"github.com/justwatchcom/gopass/config"
	"github.com/pkg/errors"
)

// Config returns this root stores config as a config struct
func (s *Store) Config() *config.Config {
	c := &config.Config{
		AskForMore:  s.askForMore,
		AutoImport:  s.autoImport,
		AutoSync:    s.autoSync,
		ClipTimeout: s.clipTimeout,
		Mounts:      make(map[string]string, len(s.mounts)),
		NoColor:     s.noColor,
		NoConfirm:   s.noConfirm,
		NoPager:     s.noPager,
		Path:        s.path,
		SafeContent: s.safeContent,
		Version:     s.version,
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
	s.askForMore = cfg.AskForMore
	s.autoImport = cfg.AutoImport
	s.autoSync = cfg.AutoSync
	s.clipTimeout = cfg.ClipTimeout
	s.noColor = cfg.NoColor
	s.noConfirm = cfg.NoConfirm
	s.noPager = cfg.NoPager
	s.path = cfg.Path
	s.safeContent = cfg.SafeContent

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

// AutoSync returns the value of auto sync
func (s *Store) AutoSync() bool {
	return s.autoSync
}

// SafeContent returns the value of safe content
func (s *Store) SafeContent() bool {
	return s.safeContent
}

// ClipTimeout returns the value of clip timeout
func (s *Store) ClipTimeout() int {
	return s.clipTimeout
}

// AskForMore returns true if generate should ask for more information
func (s *Store) AskForMore() bool {
	return s.askForMore
}

// NoPager retuns true if the user doesn't want a pager for longer output
func (s *Store) NoPager() bool {
	return s.noPager
}

// NoConfirm returns true if no recipients should be confirmed on encryption
func (s *Store) NoConfirm() bool {
	return s.noConfirm
}
