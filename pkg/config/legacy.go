package config

import "github.com/gopasspw/gopass/pkg/backend"

// Pre182 is the current config struct
type Pre182 struct {
	Path    string                        `yaml:"-"`
	Root    *Pre182StoreConfig            `yaml:"root"`
	Mounts  map[string]*Pre182StoreConfig `yaml:"mounts"`
	Version string                        `yaml:"version"`

	// Catches all undefined files and must be empty after parsing
	XXX map[string]interface{} `yaml:",inline"`
}

// Pre182StoreConfig is a per-store (root or mount) config
type Pre182StoreConfig struct {
	AskForMore     bool              `yaml:"askformore"` // ask for more data on generate
	AutoClip       bool              `yaml:"autoclip"`   // decide whether passwords are automatically copied or not
	AutoImport     bool              `yaml:"autoimport"` // import missing public keys w/o asking
	AutoSync       bool              `yaml:"autosync"`   // push to git remote after commit, pull before push if necessary
	CheckRecpHash  bool              `yaml:"check_recipient_hash"`
	ClipTimeout    int               `yaml:"cliptimeout"`    // clear clipboard after seconds
	Concurrency    int               `yaml:"concurrency"`    // allow to run multiple thread when batch processing
	EditRecipients bool              `yaml:"editrecipients"` // edit recipients when confirming
	NoColor        bool              `yaml:"nocolor"`        // do not use color when outputing text
	NoConfirm      bool              `yaml:"noconfirm"`      // do not confirm recipients when encrypting
	NoPager        bool              `yaml:"nopager"`        // do not invoke a pager to display long lists
	Notifications  bool              `yaml:"notifications"`  // enable desktop notifications
	Path           *backend.URL      `yaml:"path"`           // path to the root store
	RecipientHash  map[string]string `yaml:"recipient_hash"`
	SafeContent    bool              `yaml:"safecontent"` // avoid showing passwords in terminal
	UseSymbols     bool              `yaml:"usesymbols"`  // always use symbols when generating passwords
}

// StoreConfig returns a current StoreConfig
func (c *Pre182StoreConfig) StoreConfig() *StoreConfig {
	sc := StoreConfig(*c)
	return &sc
}

// CheckOverflow implements configer
func (c *Pre182) CheckOverflow() error {
	return checkOverflow(c.XXX, "config")
}

// Config converts the Pre140 config to the current config struct
func (c *Pre182) Config() *Config {
	cfg := &Config{
		Root:   c.Root.StoreConfig(),
		Mounts: make(map[string]*StoreConfig, len(c.Mounts)),
	}
	for k, v := range c.Mounts {
		cfg.Mounts[k] = v.StoreConfig()
	}
	return cfg
}

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
		Root:   &sc,
		Mounts: make(map[string]*StoreConfig, len(c.Mounts)),
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
		Root:   &sc,
		Mounts: make(map[string]*StoreConfig, len(c.Mounts)),
	}
	for k, v := range c.Mounts {
		subSc := sc
		subSc.Path = backend.FromPath(v)
		cfg.Mounts[k] = &subSc
	}
	return cfg
}
