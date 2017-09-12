package sub

import (
	"github.com/justwatchcom/gopass/config"
	"github.com/pkg/errors"
)

// Config returns this sub stores config as a config struct
func (s *Store) Config() *config.Config {
	c := &config.Config{
		Mounts: make(map[string]string),
		Path:   s.path,
	}
	return c
}

// UpdateConfig updates this sub-stores internal config
func (s *Store) UpdateConfig(cfg *config.Config) error {
	if cfg == nil {
		return errors.Errorf("invalid config")
	}
	s.path = cfg.Path
	// substores have no mounts

	return nil
}

// Path returns the value of path
func (s *Store) Path() string {
	return s.path
}

// Alias returns the value of alias
func (s *Store) Alias() string {
	return s.alias
}
