package root

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"

	"github.com/justwatchcom/gopass/config"
)

// Config returns this root stores config as a config struct
func (s *Store) Config() *config.Config {
	c := &config.Config{
		Mounts: make(map[string]string, len(s.mounts)),
	}
	for alias, sub := range s.mounts {
		c.Mounts[alias] = sub.Path()
	}
	c.FsckFunc = s.fsckFunc
	c.ImportFunc = s.importFunc

	os := reflect.ValueOf(s).Elem()
	oc := reflect.ValueOf(c).Elem()
	for i := 0; i < os.NumField(); i++ {
		gpArg := os.Type().Field(i).Tag.Get("gopass")
		if gpArg == "-" {
			continue
		}
		fs := os.Field(i)
		name := strings.Title(os.Type().Field(i).Name)
		fc := oc.FieldByName(name)
		if fc.Kind() != fs.Kind() {
			continue
		}
		switch fs.Kind() {
		case reflect.String:
			fc.SetString(fs.String())
		case reflect.Bool:
			fc.SetBool(fs.Bool())
		case reflect.Int:
			fc.SetInt(fs.Int())
		default:
			continue
		}
	}

	// trick "unused", we need those when mounting a new sub-store
	_ = s.alwaysTrust
	_ = s.fsckFunc
	_ = s.loadKeys
	_ = s.noColor
	_ = s.persistKeys
	_ = s.version

	return c
}

// UpdateConfig updates this root-stores internal config and propagates
// those changes to all substores
func (s *Store) UpdateConfig(cfg *config.Config) error {
	if cfg == nil {
		return fmt.Errorf("invalid config")
	}

	s.fsckFunc = cfg.FsckFunc
	s.importFunc = cfg.ImportFunc

	os := reflect.ValueOf(s).Elem()
	oc := reflect.ValueOf(cfg).Elem()
	for i := 0; i < os.NumField(); i++ {
		gpArg := os.Type().Field(i).Tag.Get("gopass")
		if gpArg == "-" {
			continue
		}
		fs := os.Field(i)
		name := strings.Title(os.Type().Field(i).Name)
		fc := oc.FieldByName(name)
		if fc.Kind() != fs.Kind() {
			continue
		}
		if !fs.CanAddr() {
			continue
		}
		// Acording to the "rules of reflect" fields obtained through unexported
		// fields can not be updated. The following line creates a writeable
		// copy at the exact same location to make it writeable.
		// see https://stackoverflow.com/a/43918797/218846
		fs = reflect.NewAt(fs.Type(), unsafe.Pointer(fs.UnsafeAddr())).Elem()
		switch fc.Kind() {
		case reflect.String:
			fs.SetString(fc.String())
		case reflect.Bool:
			fs.SetBool(fc.Bool())
		case reflect.Int:
			fs.SetInt(fc.Int())
		default:
			continue
		}
	}

	// add any missing mounts
	for alias, path := range cfg.Mounts {
		if _, found := s.mounts[alias]; !found {
			if err := s.addMount(alias, path); err != nil {
				return err
			}
		}
	}

	// propagate any config changes to our substores
	if s.store != nil {
		if err := s.store.UpdateConfig(cfg); err != nil {
			return err
		}
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
