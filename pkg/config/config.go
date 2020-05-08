package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/pkg/errors"
)

var (
	// ErrConfigNotFound is returned on load if the config was not found
	ErrConfigNotFound = errors.Errorf("config not found")
	// ErrConfigNotParsed is returned on load if the config could not be decoded
	ErrConfigNotParsed = errors.Errorf("config not parseable")
	debug              = false
)

func init() {
	if gdb := os.Getenv("GOPASS_DEBUG"); gdb != "" {
		debug = true
	}
}

// Config is the current config struct
type Config struct {
	Path   string                  `yaml:"-"`
	Root   *StoreConfig            `yaml:"root"`
	Mounts map[string]*StoreConfig `yaml:"mounts"`

	// Catches all undefined files and must be empty after parsing
	XXX map[string]interface{} `yaml:",inline"`
}

// New creates a new config with sane default values
func New() *Config {
	return &Config{
		Path: configLocation(),
		Root: &StoreConfig{
			AskForMore:    false,
			AutoClip:      false,
			AutoImport:    true,
			AutoSync:      true,
			ClipTimeout:   45,
			Concurrency:   1,
			ExportKeys:    true,
			NoColor:       false,
			NoConfirm:     false,
			NoPager:       false,
			SafeContent:   false,
			UseSymbols:    false,
			Notifications: true,
		},
		Mounts: make(map[string]*StoreConfig),
	}
}

// CheckOverflow implements configer. It will check for any extra config values not
// handled by the current struct
func (c *Config) CheckOverflow() error {
	return checkOverflow(c.XXX, "config")
}

// Config will return a current config
func (c *Config) Config() *Config {
	return c
}

// SetConfigValue will try to set the given key to the value in the config struct
func (c *Config) SetConfigValue(mount, key, value string) error {
	if mount == "" {
		if err := c.Root.SetConfigValue(key, value); err != nil {
			return err
		}
		return c.Save()
	}

	if sc, found := c.Mounts[mount]; found {
		if err := sc.SetConfigValue(key, value); err != nil {
			return err
		}
	}
	return c.Save()
}

func (c *Config) checkDefaults() error {
	if c == nil {
		return fmt.Errorf("config is nil")
	}
	if c.Root == nil {
		c.Root = &StoreConfig{}
	}
	if err := c.Root.checkDefaults(); err != nil {
		return err
	}
	for _, sc := range c.Mounts {
		if err := sc.checkDefaults(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Config) String() string {
	mounts := ""
	keys := make([]string, 0, len(c.Mounts))
	for alias := range c.Mounts {
		keys = append(keys, alias)
	}
	sort.Strings(keys)

	for _, alias := range keys {
		sc := c.Mounts[alias]
		mounts += alias + "=>" + sc.String()
	}
	return fmt.Sprintf("Config[Root:%s,Mounts(%s)]", c.Root.String(), mounts)
}

// Directory returns the directory this config is using
func (c *Config) Directory() string {
	return filepath.Dir(c.Path)
}

// GetRecipientHash returns the recipients hash for the given store and file
func (c *Config) GetRecipientHash(alias, name string) string {
	if alias == "" {
		return c.Root.RecipientHash[name]
	}
	if sc, found := c.Mounts[alias]; found && sc != nil {
		return sc.RecipientHash[name]
	}
	return ""
}

// SetRecipientHash will set and save the recipient hash for the given store
// and file
func (c *Config) SetRecipientHash(alias, name, value string) error {
	if alias == "" {
		c.Root.setRecipientHash(name, value)
	} else {
		if sc, found := c.Mounts[alias]; found && sc != nil {
			sc.setRecipientHash(name, value)
		}
	}

	return c.Save()
}

// CheckRecipientHash returns true if we should report/fail on any
// recipient hash errors for this store
func (c *Config) CheckRecipientHash(alias string) bool {
	if alias == "" {
		return c.Root.CheckRecpHash
	}
	if sc, found := c.Mounts[alias]; found && sc != nil {
		return sc.CheckRecpHash
	}
	return false
}
