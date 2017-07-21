package sub

import (
	"fmt"

	"github.com/justwatchcom/gopass/config"
)

// Config returns this sub stores config as a config struct
func (s *Store) Config() *config.Config {
	c := &config.Config{
		AutoSync:   s.autoSync,
		AutoImport: s.autoImport,
		FsckFunc:   s.fsckFunc,
		ImportFunc: s.importFunc,
		Mounts:     make(map[string]string),
		Path:       s.path,
	}
	return c
}

// UpdateConfig updates this sub-stores internal config
func (s *Store) UpdateConfig(cfg *config.Config) error {
	if cfg == nil {
		return fmt.Errorf("invalid config")
	}
	s.autoImport = cfg.AutoImport
	s.autoSync = cfg.AutoSync
	s.fsckFunc = cfg.FsckFunc
	s.importFunc = cfg.ImportFunc
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
