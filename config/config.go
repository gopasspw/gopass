package config

import (
	"os"
	"strings"

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
			AskForMore:  false,
			AutoImport:  true,
			AutoSync:    true,
			ClipTimeout: 45,
			NoConfirm:   false,
			NoPager:     false,
			SafeContent: false,
			UseSymbols:  false,
			NoColor:     false,
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
	c.Root.Path = stripURL(c.Root.Path)
	for k := range c.Mounts {
		c.Mounts[k].Path = stripURL(c.Mounts[k].Path)
	}
	return c
}

func stripURL(in string) string {
	pat := "file://"
	if !strings.Contains(in, pat) {
		return in
	}
	return in[strings.Index(in, pat)+len(pat):]
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
