package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/justwatchcom/gopass/store"
	"github.com/justwatchcom/gopass/utils/fsutil"
	"github.com/pkg/errors"
)

var (
	// ErrConfigNotFound is returned on load if the config was not found
	ErrConfigNotFound = errors.Errorf("config not found")
	// ErrConfigNotParsed is returned on load if the config could not be decoded
	ErrConfigNotParsed = errors.Errorf("config not parseable")
)

// Config is the gopass config structure
type Config struct {
	AskForMore  bool                 `json:"askformore"`  // ask for more data on generate
	AutoImport  bool                 `json:"autoimport"`  // import missing public keys w/o asking
	AutoSync    bool                 `json:"autosync"`    // push to git remote after commit, pull before push if necessary
	ClipTimeout int                  `json:"cliptimeout"` // clear clipboard after seconds
	Debug       bool                 `json:"-"`
	FsckFunc    store.FsckCallback   `json:"-"`
	ImportFunc  store.ImportCallback `json:"-"`
	Mounts      map[string]string    `json:"mounts,omitempty"`
	NoColor     bool                 `json:"-"`
	NoPager     bool                 `json:"-"`
	NoConfirm   bool                 `json:"noconfirm"`   // do not confirm recipients when encrypting
	Path        string               `json:"path"`        // path to the root store
	SafeContent bool                 `json:"safecontent"` // avoid showing passwords in terminal
	Version     string               `json:"version"`
}

// New creates a new config with sane default values
func New() *Config {
	return &Config{
		AskForMore:  false,
		AutoImport:  true,
		AutoSync:    true,
		ClipTimeout: 45,
		Mounts:      make(map[string]string),
		NoConfirm:   false,
		SafeContent: false,
		Version:     "",
	}
}

// ConfigMap returns a map of stringified config values for easy printing
func (c *Config) ConfigMap() map[string]string {
	m := make(map[string]string, 20)
	o := reflect.ValueOf(c).Elem()
	for i := 0; i < o.NumField(); i++ {
		jsonArg := o.Type().Field(i).Tag.Get("json")
		if jsonArg == "" || jsonArg == "-" {
			continue
		}
		f := o.Field(i)
		strVal := ""
		switch f.Kind() {
		case reflect.String:
			strVal = f.String()
		case reflect.Bool:
			strVal = fmt.Sprintf("%t", f.Bool())
		case reflect.Int:
			strVal = fmt.Sprintf("%d", f.Int())
		default:
			continue
		}
		m[jsonArg] = strVal
	}
	return m
}

// SetConfigValue will try to set the given key to the value in the config struct
func (c *Config) SetConfigValue(key, value string) error {
	if key == "version" {
		return errors.Errorf("Can not change version")
	}
	if key != "path" {
		value = strings.ToLower(value)
	}
	o := reflect.ValueOf(c).Elem()
	for i := 0; i < o.NumField(); i++ {
		jsonArg := o.Type().Field(i).Tag.Get("json")
		if jsonArg == "" || jsonArg == "-" {
			continue
		}
		if jsonArg != key {
			continue
		}
		f := o.Field(i)
		switch f.Kind() {
		case reflect.String:
			f.SetString(value)
		case reflect.Bool:
			if value == "true" {
				f.SetBool(true)
			} else if value == "false" {
				f.SetBool(false)
			} else {
				return errors.Errorf("No a bool: %s", value)
			}
		case reflect.Int:
			iv, err := strconv.Atoi(value)
			if err != nil {
				return errors.Wrapf(err, "failed to convert '%s' to int", value)
			}
			f.SetInt(int64(iv))
		default:
			continue
		}
	}
	return c.Save()
}

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
		if gdb := os.Getenv("GOPASS_DEBUG"); gdb == "true" {
			fmt.Printf("[DEBUG] Loaded config from %s: %+v\n", l, cfg)
		}
		return cfg
	}
	cfg := New()
	cfg.Path = PwStoreDir("")
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
	cfg := &Config{
		AutoSync: true,
	}
	err = yaml.Unmarshal(buf, &cfg)
	if err != nil {
		fmt.Printf("Error reading config from %s: %s\n", cf, err)
		return nil, ErrConfigNotParsed
	}
	return cfg, nil
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
	return nil
}

// configLocation returns the location of the config file. Either reading from
// GOPASS_CONFIG or using the default location (~/.gopass.yml)
func configLocation() string {
	if cf := os.Getenv("GOPASS_CONFIG"); cf != "" {
		return cf
	}
	if xch := os.Getenv("XDG_CONFIG_HOME"); xch != "" {
		return filepath.Join(xch, "gopass", "config.yml")
	}
	return filepath.Join(os.Getenv("HOME"), ".config", "gopass", "config.yml")
}

// configLocations returns the possible locations of gopass config files,
// in decreasing priority
func configLocations() []string {
	l := []string{}
	if cf := os.Getenv("GOPASS_CONFIG"); cf != "" {
		l = append(l, cf)
	}
	if xch := os.Getenv("XDG_CONFIG_HOME"); xch != "" {
		l = append(l, filepath.Join(xch, "gopass", "config.yml"))
	}
	l = append(l, filepath.Join(os.Getenv("HOME"), ".config", "gopass", "config.yml"))
	l = append(l, filepath.Join(os.Getenv("HOME"), ".gopass.yml"))
	return l
}

// PwStoreDir reads the password store dir from the environment
// or returns the default location ~/.password-store if the env is
// not set
func PwStoreDir(mount string) string {
	if mount != "" {
		return fsutil.CleanPath(filepath.Join(os.Getenv("HOME"), ".password-store-"+strings.Replace(mount, string(filepath.Separator), "-", -1)))
	}
	if d := os.Getenv("PASSWORD_STORE_DIR"); d != "" {
		return fsutil.CleanPath(d)
	}
	return os.Getenv("HOME") + "/.password-store"
}
