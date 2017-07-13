package sub

import (
	"fmt"

	"github.com/justwatchcom/gopass/config"
)

// Config returns this sub stores config as a config struct
func (s *Store) Config() *config.Config {
	c := &config.Config{
		AlwaysTrust:     s.alwaysTrust,
		AutoImport:      s.autoImport,
		AutoPull:        s.autoPull,
		AutoPush:        s.autoPush,
		CheckRecipients: s.checkRecipients,
		Debug:           s.debug,
		FsckFunc:        s.fsckFunc,
		ImportFunc:      s.importFunc,
		LoadKeys:        s.loadKeys,
		Mounts:          make(map[string]string),
		Path:            s.path,
		PersistKeys:     s.persistKeys,
	}
	return c
}

// UpdateConfig updates this sub-stores internal config
func (s *Store) UpdateConfig(cfg *config.Config) error {
	if cfg == nil {
		return fmt.Errorf("invalid config")
	}
	s.alwaysTrust = cfg.AlwaysTrust
	s.autoImport = cfg.AutoImport
	s.autoPull = cfg.AutoPull
	s.autoPush = cfg.AutoPush
	s.checkRecipients = cfg.CheckRecipients
	s.debug = cfg.Debug
	s.fsckFunc = cfg.FsckFunc
	s.importFunc = cfg.ImportFunc
	s.loadKeys = cfg.LoadKeys
	s.path = cfg.Path
	s.persistKeys = cfg.PersistKeys

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
