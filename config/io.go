package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/justwatchcom/gopass/utils/fsutil"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// Load will try to load the config from one of the default locations
func Load() *Config {
	for _, l := range configLocations() {
		cfg, err := load(l)
		if err == ErrConfigNotFound {
			continue
		}
		if err != nil {
			panic(err)
		}
		if debug {
			fmt.Printf("[DEBUG] Loaded config from %s: %+v\n", l, cfg)
		}
		return cfg
	}
	cfg := New()
	cfg.Root.Path = PwStoreDir("")
	cfg.checkDefaults()
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
		fmt.Printf("Error reading config from %s: %s\n", cf, err)
		return nil, ErrConfigNotFound
	}

	cfg, err := decode(buf)
	if err != nil {
		fmt.Printf("Error reading config from %s: %s\n", cf, err)
		return nil, ErrConfigNotParsed
	}
	if cfg.Mounts == nil {
		cfg.Mounts = make(map[string]*StoreConfig)
	}
	return cfg, nil
}

func checkOverflow(m map[string]interface{}, section string) error {
	if len(m) < 1 {
		return nil
	}

	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	return errors.Errorf("unknown fields in %s: %+v", section, keys)
}

type configer interface {
	Config() *Config
	CheckOverflow() error
}

func decode(buf []byte) (*Config, error) {
	cfgs := []configer{
		&Config{
			Root: &StoreConfig{
				AutoSync: true,
			},
		},
		&Pre140{
			AutoSync: true,
		},
		&Pre130{},
	}
	for _, cfg := range cfgs {
		if err := yaml.Unmarshal(buf, cfg); err != nil {
			continue
		}
		if err := cfg.CheckOverflow(); err != nil {
			if debug {
				fmt.Printf("[DEBUG] Extra elements in config: %s\n", err)
			}
			continue
		}
		if debug {
			fmt.Printf("[DEBUG] Loaded config: %+v\n", cfg)
		}
		return cfg.Config(), nil
	}
	return nil, ErrConfigNotParsed
}

// Save saves the config
func (c *Config) Save() error {
	c.checkDefaults()
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
	if debug {
		fmt.Printf("[DEBUG] Saved config to %s: %+v\n", cfgLoc, c)
	}
	return nil
}
