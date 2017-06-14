package sub

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"

	"github.com/justwatchcom/gopass/config"
)

// Config returns this sub stores config as a config struct
func (s *Store) Config() *config.Config {
	c := &config.Config{
		Mounts: make(map[string]string),
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

	return c
}

// UpdateConfig updates this sub-stores internal config
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
