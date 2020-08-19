package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/gopasspw/gopass/internal/debug"
	"github.com/gopasspw/gopass/pkg/fsutil"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// LoadWithFallback will try to load the config from one of the default locations
// TODO(2.x) This method is DEPRECATED
func LoadWithFallback() *Config {
	for _, l := range configLocations() {
		if cfg := loadConfig(l); cfg != nil {
			return cfg
		}
	}
	return loadDefault()
}

// Load will load the config from the default location or return a default config
func Load() *Config {
	if cfg := loadConfig(configLocation()); cfg != nil {
		return cfg
	}
	return loadDefault()
}

func loadConfig(l string) *Config {
	debug.Log("Trying to load config from %s", l)
	cfg, err := load(l)
	if err == ErrConfigNotFound {
		return nil
	}
	if err != nil {
		return nil
	}
	debug.Log("Loaded config from %s: %+v", l, cfg)
	return cfg
}

func loadDefault() *Config {
	cfg := New()
	cfg.Path = PwStoreDir("")
	debug.Log("Loaded default config: %+v", cfg)
	return cfg
}

func load(cf string) (*Config, error) {
	// deliberately using os.Stat here, a symlinked
	// config is OK
	if _, err := os.Stat(cf); err != nil {
		return nil, ErrConfigNotFound
	}
	buf, err := ioutil.ReadFile(cf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading config from %s: %s\n", cf, err)
		return nil, ErrConfigNotFound
	}

	cfg, err := decode(buf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading config from %s: %s\n", cf, err)
		return nil, ErrConfigNotParsed
	}
	if cfg.Mounts == nil {
		cfg.Mounts = make(map[string]string)
	}
	cfg.configPath = cf
	return cfg, nil
}

func checkOverflow(m map[string]interface{}) error {
	if len(m) < 1 {
		return nil
	}

	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return errors.Errorf("unknown fields: %+v", keys)
}

type configer interface {
	Config() *Config
	CheckOverflow() error
}

func decode(buf []byte) (*Config, error) {
	cfgs := []configer{
		&Config{
			ExportKeys: true,
		},
		&Pre193{
			Root: &Pre193StoreConfig{},
		},
		&Pre182{
			Root: &Pre182StoreConfig{},
		},
		&Pre140{},
		&Pre130{},
	}
	for i, cfg := range cfgs {
		debug.Log("Trying to unmarshal config into %T", cfg)
		if err := yaml.Unmarshal(buf, cfg); err != nil {
			debug.Log("Loading config %T failed: %s", cfg, err)
			continue
		}
		if err := cfg.CheckOverflow(); err != nil {
			debug.Log("Extra elements in config: %s", err)
			if i == 0 {
				fmt.Fprintf(os.Stderr, "Failed to load config %T. %s\n", cfg, err)
			}
			continue
		}
		debug.Log("Loaded config: %T: %+v", cfg, cfg)
		conf := cfg.Config()
		if i > 0 {
			debug.Log("Loaded legacy config. Should rewrite config.")
		}
		return conf, nil
	}
	return nil, ErrConfigNotParsed
}

// Save saves the config
func (c *Config) Save() error {
	buf, err := yaml.Marshal(c)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal YAML")
	}

	cfgLoc := configLocation()
	cfgDir := filepath.Dir(cfgLoc)
	if !fsutil.IsDir(cfgDir) {
		if err := os.MkdirAll(cfgDir, 0700); err != nil {
			return errors.Wrapf(err, "failed to create dir '%s'", cfgDir)
		}
	}
	if err := ioutil.WriteFile(cfgLoc, buf, 0600); err != nil {
		return errors.Wrapf(err, "failed to write config file to '%s'", cfgLoc)
	}
	debug.Log("Saved config to %s: %+v\n", cfgLoc, c)
	return nil
}
