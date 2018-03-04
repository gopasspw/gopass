package config

import (
	"fmt"
	"os"
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
	Root    *StoreConfig            `yaml:"root"`
	Mounts  map[string]*StoreConfig `yaml:"mounts"`
	Version string                  `yaml:"version"`

	// Catches all undefined files and must be empty after parsing
	XXX map[string]interface{} `yaml:",inline"`
}

// New creates a new config with sane default values
func New() *Config {
	return &Config{
		Root: &StoreConfig{
			AskForMore:    false,
			AutoImport:    true,
			AutoSync:      true,
			ClipTimeout:   45,
			CryptoBackend: "gpg",
			NoColor:       false,
			NoConfirm:     false,
			NoPager:       false,
			SafeContent:   false,
			StoreBackend:  "fs",
			SyncBackend:   "git",
			UseSymbols:    false,
			Notifications: true,
		},
		Mounts:  make(map[string]*StoreConfig),
		Version: "",
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

func (c *Config) checkDefaults() {
	if c == nil {
		return
	}
	if c.Root == nil {
		c.Root = &StoreConfig{}
	}
	if c.Root.CryptoBackend == "" {
		c.Root.CryptoBackend = "gpg"
	}
	if c.Root.SyncBackend == "" {
		c.Root.SyncBackend = "git"
	}
	if c.Root.StoreBackend == "" {
		c.Root.StoreBackend = "fs"
	}
	for _, sc := range c.Mounts {
		if sc.CryptoBackend == "" {
			sc.CryptoBackend = "gpg"
		}
		if sc.SyncBackend == "" {
			sc.SyncBackend = "git"
		}
		if sc.StoreBackend == "" {
			sc.StoreBackend = "fs"
		}
	}
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
	return fmt.Sprintf("Config[Root:%s,Mounts(%s),Version:%s]", c.Root.String(), mounts, c.Version)
}
