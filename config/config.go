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
	"github.com/justwatchcom/gopass/fsutil"
	"github.com/justwatchcom/gopass/store"
)

// Config is the gopass config structure
type Config struct {
	AlwaysTrust bool                 `json:"alwaystrust"` // always trust public keys when encrypting
	AskForMore  bool                 `json:"askformore"`  // ask for more data on generate
	AutoImport  bool                 `json:"autoimport"`  // import missing public keys w/o asking
	AutoPull    bool                 `json:"autopull"`    // pull from git before push
	AutoPush    bool                 `json:"autopush"`    // push to git remote after commit
	ClipTimeout int                  `json:"cliptimeout"` // clear clipboard after seconds
	Debug       bool                 `json:"debug"`       // enable debug output
	FsckFunc    store.FsckCallback   `json:"-"`
	ImportFunc  store.ImportCallback `json:"-"`
	LoadKeys    bool                 `json:"loadkeys"` // load missing keys from store
	Mounts      map[string]string    `json:"mounts,omitempty"`
	NoColor     bool                 `json:"nocolor"`     // disable colors in output
	NoConfirm   bool                 `json:"noconfirm"`   // do not confirm recipients when encrypting
	Path        string               `json:"path"`        // path to the root store
	PersistKeys bool                 `json:"persistkeys"` // store recipient keys in store
	SafeContent bool                 `json:"safecontent"` // avoid showing passwords in terminal
	Version     string               `json:"version"`
}

// New creates a new config with sane default values
func New() *Config {
	return &Config{
		AlwaysTrust: true,
		AskForMore:  false,
		AutoImport:  true,
		AutoPull:    true,
		AutoPush:    true,
		ClipTimeout: 45,
		Debug:       false,
		LoadKeys:    true,
		Mounts:      make(map[string]string),
		NoColor:     false,
		NoConfirm:   false,
		PersistKeys: true,
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
		return fmt.Errorf("Can not change version")
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
				return fmt.Errorf("No a bool: %s", value)
			}
		case reflect.Int:
			iv, err := strconv.Atoi(value)
			if err != nil {
				return err
			}
			f.SetInt(int64(iv))
		default:
			continue
		}
	}
	return c.Save()
}

// Load will try to load the config from one of the default locations
func Load() (*Config, error) {
	for _, l := range configLocations() {
		if cfg, err := load(l); err == nil {
			if gdb := os.Getenv("GOPASS_DEBUG"); gdb == "true" {
				fmt.Printf("[DEBUG] Loaded config from %s: %+v\n", l, cfg)
			}
			return cfg, err
		}
	}
	return nil, fmt.Errorf("no config found")
}

func load(cf string) (*Config, error) {
	// deliberately using os.Stat here, a symlinked
	// config is OK
	if _, err := os.Stat(cf); err != nil {
		return nil, err
	}
	buf, err := ioutil.ReadFile(cf)
	if err != nil {
		fmt.Printf("Error reading config from %s: %s\n", cf, err)
		return nil, err
	}
	cfg := &Config{}
	err = yaml.Unmarshal(buf, &cfg)
	if err != nil {
		fmt.Printf("Error reading config from %s: %s\n", cf, err)
		return nil, err
	}
	return cfg, nil
}

// Save saves the config
func (c *Config) Save() error {
	buf, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	cfgLoc := configLocation()
	cfgDir := filepath.Dir(cfgLoc)
	if !fsutil.IsDir(cfgDir) {
		if err := os.MkdirAll(cfgDir, 0700); err != nil {
			return err
		}
	}
	if err := ioutil.WriteFile(cfgLoc, buf, 0600); err != nil {
		return err
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
