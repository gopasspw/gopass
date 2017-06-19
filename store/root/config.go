package root

import (
	"fmt"

	"github.com/justwatchcom/gopass/config"
)

// Config returns this root stores config as a config struct
func (s *Store) Config() *config.Config {
	c := &config.Config{
		AlwaysTrust: s.alwaysTrust,
		AskForMore:  s.askForMore,
		AutoImport:  s.autoImport,
		AutoPull:    s.autoPull,
		AutoPush:    s.autoPush,
		ClipTimeout: s.clipTimeout,
		Debug:       s.debug,
		LoadKeys:    s.loadKeys,
		Mounts:      make(map[string]string, len(s.mounts)),
		NoColor:     s.noColor,
		NoConfirm:   s.noConfirm,
		Path:        s.path,
		PersistKeys: s.persistKeys,
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
func (s *Store) UpdateConfig(cfg *config.Config) error {
	if cfg == nil {
		return fmt.Errorf("invalid config")
	}
	s.alwaysTrust = cfg.AlwaysTrust
	s.askForMore = cfg.AskForMore
	s.autoImport = cfg.AutoImport
	s.autoPull = cfg.AutoPull
	s.autoPush = cfg.AutoPush
	s.debug = cfg.Debug
	s.clipTimeout = cfg.ClipTimeout
	s.loadKeys = cfg.LoadKeys
	s.noColor = cfg.NoColor
	s.noConfirm = cfg.NoConfirm
	s.path = cfg.Path
	s.persistKeys = cfg.PersistKeys
	s.safeContent = cfg.SafeContent

	// add any missing mounts
	for alias, path := range cfg.Mounts {
		if _, found := s.mounts[alias]; !found {
			if err := s.addMount(alias, path); err != nil {
				return err
			}
		}
	}

	// propagate any config changes to our substores
	if err := s.store.UpdateConfig(cfg); err != nil {
		return err
	}
	for _, sub := range s.mounts {
		if err := sub.UpdateConfig(cfg); err != nil {
			return err
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

// NoConfirm returns true if no recipients should be confirmed on encryption
func (s *Store) NoConfirm() bool {
	return s.noConfirm
}

// AutoPush returns the value of auto push
func (s *Store) AutoPush() bool {
	return s.autoPush
}

// AutoPull returns the value of auto pull
func (s *Store) AutoPull() bool {
	return s.autoPull
}

// AutoImport returns the value of auto import
func (s *Store) AutoImport() bool {
	return s.autoImport
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
