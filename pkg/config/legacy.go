package config

import "github.com/gopasspw/gopass/pkg/backend"

// Pre140 is the gopass config structure before version 1.4.0
type Pre140 struct {
	AskForMore  bool              `yaml:"askformore"`  // ask for more data on generate
	AutoImport  bool              `yaml:"autoimport"`  // import missing public keys w/o asking
	AutoSync    bool              `yaml:"autosync"`    // push to git remote after commit, pull before push if necessary
	ClipTimeout int               `yaml:"cliptimeout"` // clear clipboard after seconds
	Mounts      map[string]string `yaml:"mounts,omitempty"`
	NoConfirm   bool              `yaml:"noconfirm"`   // do not confirm recipients when encrypting
	Path        string            `yaml:"path"`        // path to the root store
	SafeContent bool              `yaml:"safecontent"` // avoid showing passwords in terminal
	Version     string            `yaml:"version"`

	// Catches all undefined files and must be empty after parsing
	XXX map[string]interface{} `yaml:",inline"`
}

// CheckOverflow implements configer
func (c *Pre140) CheckOverflow() error {
	return checkOverflow(c.XXX, "config")
}

// Config converts the Pre140 config to the current config struct
func (c *Pre140) Config() *Config {
	sc := StoreConfig{
		AskForMore:  c.AskForMore,
		AutoImport:  c.AutoImport,
		AutoSync:    c.AutoSync,
		ClipTimeout: c.ClipTimeout,
		NoConfirm:   c.NoConfirm,
		Path:        backend.FromPath(c.Path),
		SafeContent: c.SafeContent,
	}
	cfg := &Config{
		Root:    &sc,
		Mounts:  make(map[string]*StoreConfig, len(c.Mounts)),
		Version: c.Version,
	}
	for k, v := range c.Mounts {
		subSc := sc
		subSc.Path = backend.FromPath(v)
		cfg.Mounts[k] = &subSc
	}
	return cfg
}

// Pre130 is the gopass config structure before version 1.3.0. Not all fields were
// available between 1.0.0 and 1.3.0, but this struct should cover all of them.
type Pre130 struct {
	AlwaysTrust bool              `yaml:"alwaystrust"` // always trust public keys when encrypting
	AskForMore  bool              `yaml:"askformore"`  // ask for more data on generate
	AutoImport  bool              `yaml:"autoimport"`  // import missing public keys w/o asking
	AutoPull    bool              `yaml:"autopull"`    // pull from git before push
	AutoPush    bool              `yaml:"autopush"`    // push to git remote after commit
	ClipTimeout int               `yaml:"cliptimeout"` // clear clipboard after seconds
	Debug       bool              `yaml:"debug"`       // enable debug output
	LoadKeys    bool              `yaml:"loadkeys"`    // load missing keys from store
	Mounts      map[string]string `yaml:"mounts,omitempty"`
	NoColor     bool              `yaml:"nocolor"`     // disable colors in output
	NoConfirm   bool              `yaml:"noconfirm"`   // do not confirm recipients when encrypting
	Path        string            `yaml:"path"`        // path to the root store
	PersistKeys bool              `yaml:"persistkeys"` // store recipient keys in store
	SafeContent bool              `yaml:"safecontent"` // avoid showing passwords in terminal
	Version     string            `yaml:"version"`

	// Catches all undefined files and must be empty after parsing
	XXX map[string]interface{} `yaml:",inline"`
}

// CheckOverflow implements configer
func (c *Pre130) CheckOverflow() error {
	return checkOverflow(c.XXX, "config")
}

// Config converts the Pre130 config to the current config struct
func (c *Pre130) Config() *Config {
	sc := StoreConfig{
		AskForMore:  c.AskForMore,
		AutoImport:  c.AutoImport,
		AutoSync:    c.AutoPull && c.AutoPush,
		ClipTimeout: c.ClipTimeout,
		NoConfirm:   c.NoConfirm,
		Path:        backend.FromPath(c.Path),
		SafeContent: c.SafeContent,
	}
	cfg := &Config{
		Root:    &sc,
		Mounts:  make(map[string]*StoreConfig, len(c.Mounts)),
		Version: c.Version,
	}
	for k, v := range c.Mounts {
		subSc := sc
		subSc.Path = backend.FromPath(v)
		cfg.Mounts[k] = &subSc
	}
	return cfg
}
