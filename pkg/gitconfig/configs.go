package gitconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gopasspw/gopass/internal/set"
	"github.com/gopasspw/gopass/pkg/appdir"
	"github.com/gopasspw/gopass/pkg/debug"
)

// Configs is a container for a config "view" that is composed of several different
// config objects. The intention is for the ones with a wider scope to provide defaults
// so those with a more narrow scope then only have to override what they are interested in.
type Configs struct {
	Preset   *Config
	system   *Config
	global   *Config
	local    *Config
	worktree *Config
	env      *Config
	workdir  string

	SystemConfig   string
	GlobalConfig   string
	LocalConfig    string
	WorktreeConfig string
	EnvPrefix      string
	NoWrites       bool
}

func New() *Configs {
	return &Configs{
		system: &Config{
			readonly: true,
		},
		global: &Config{
			path: globalConfigFile(),
		},
		local:    &Config{},
		worktree: &Config{},
		env: &Config{
			noWrites: true,
		},

		SystemConfig:   systemConfig,
		GlobalConfig:   globalConfig,
		LocalConfig:    localConfig,
		WorktreeConfig: worktreeConfig,
		EnvPrefix:      envPrefix,
	}
}

// Reload will reload the config(s) from disk.
func (cs *Configs) Reload() {
	cs.LoadAll(cs.workdir)
}

// LoadAll tries to load all known config files. Missing or invalid files are
// silently ignored. It never fails. The workdir is optional. If non-empty
// this method will try to load a local config from this location.
func (cs *Configs) LoadAll(workdir string) *Configs {
	cs.workdir = workdir

	debug.Log("Loading gitconfigs for %+v ...", cs)

	// load the system config, if any
	if os.Getenv(cs.EnvPrefix+"_NOSYSTEM") == "" {
		c, err := LoadConfig(cs.SystemConfig)
		if err != nil {
			debug.Log("failed to load system config: %s", err)
		} else {
			debug.Log("loaded system config from %s", cs.SystemConfig)
			cs.system = c
			// the system config should generally not be written from gopass.
			// in almost any scenario gopass shouldn't have write access
			// and even if it does we shouldn't accidentially change it.
			// It's for operators and package mainatiners.
			cs.system.readonly = true
		}
	}

	// load the "global" (per user) config, if any
	cs.loadGlobalConfigs()
	cs.global.noWrites = cs.NoWrites

	// load the local config, if any
	if workdir != "" {
		localConfigPath := filepath.Join(workdir, cs.LocalConfig)
		c, err := LoadConfig(localConfigPath)
		if err != nil {
			debug.Log("failed to load local config from %s: %s", localConfigPath, err)
			// set the path just in case we want to modify / write to it later
			cs.local.path = localConfigPath
		} else {
			debug.Log("loaded local config from %s", localConfigPath)
			cs.local = c
		}
	}
	cs.local.noWrites = cs.NoWrites

	// load the worktree config, if any
	if workdir != "" {
		worktreeConfigPath := filepath.Join(workdir, cs.WorktreeConfig)
		c, err := LoadConfig(worktreeConfigPath)
		if err != nil {
			debug.Log("failed to load worktree config from %s: %s", worktreeConfigPath, err)
			// set the path just in case we want to modify / write to it later
			cs.worktree.path = worktreeConfigPath
		} else {
			debug.Log("loaded local config from %s", worktreeConfigPath)
			cs.worktree = c
		}
	}
	cs.worktree.noWrites = cs.NoWrites

	// load any env vars
	cs.env = LoadConfigFromEnv(cs.EnvPrefix)

	return cs
}

func globalConfigFile() string {
	// $XDG_CONFIG_HOME/git/config
	return filepath.Join(appdir.UserConfig(), "config")
}

// loadGlobalConfigs will try to load the per-user (Git calls them "global") configs.
// Since we might need to try different locations but only want to use the first one
// it's easier to handle this in it's own method.
func (c *Configs) loadGlobalConfigs() string {
	locs := []string{
		globalConfigFile(),
	}

	if c.GlobalConfig != "" {
		// ~/.gitconfig
		locs = append(locs, filepath.Join(appdir.UserHome(), c.GlobalConfig))
	}

	for _, p := range locs {
		// GlobalConfig might be set to an empty string to disable it
		// and instead of the XDG_CONFIG_HOME path only.
		if p == "" {
			continue
		}
		cfg, err := LoadConfig(p)
		if err != nil {
			debug.Log("failed to load global config from %s", p)

			continue
		}

		debug.Log("loaded global config from %s", p)
		c.global = cfg

		return p
	}

	debug.Log("no global config found")

	// set the path in case we want to write to it (create it) later
	c.global = &Config{
		path: globalConfigFile(),
	}

	return ""
}

// HasGlobalConfig indicates if a per-user config can be found.
func (c *Configs) HasGlobalConfig() bool {
	return c.loadGlobalConfigs() != ""
}

// Get returns the value for the given key from the first location that is found.
// Lookup order: env, worktree, local, global, system and presets.
func (c *Configs) Get(key string) string {
	for _, cfg := range []*Config{
		c.env,
		c.worktree,
		c.local,
		c.global,
		c.system,
		c.Preset,
	} {
		if cfg == nil || cfg.vars == nil {
			continue
		}
		if v, found := cfg.vars[key]; found {
			return v
		}
	}

	debug.Log("no value for %s found", key)

	return ""
}

// GetGlobal specifically ask the per-user (global) config for a key.
func (c *Configs) GetGlobal(key string) string {
	if c.global == nil {
		return ""
	}

	if v, found := c.global.vars[key]; found {
		return v
	}

	debug.Log("no value for %s found", key)

	return ""
}

// GetLocal specifically asks the per-directory (local) config for a key.
func (c *Configs) GetLocal(key string) string {
	if c.local == nil {
		return ""
	}

	if v, found := c.local.vars[key]; found {
		return v
	}

	debug.Log("no value for %s found", key)

	return ""
}

// IsSet returns true if this key is set in any of our configs.
func (c *Configs) IsSet(key string) bool {
	for _, cfg := range []*Config{
		c.env,
		c.worktree,
		c.local,
		c.global,
		c.system,
		c.Preset,
	} {
		if cfg.IsSet(key) {
			return true
		}
	}

	return false
}

// SetLocal sets (or adds) a key only in the per-directory (local) config.
func (c *Configs) SetLocal(key, value string) error {
	if c.local == nil {
		if c.workdir == "" {
			return fmt.Errorf("no workdir set")
		}
		c.local = &Config{
			path: filepath.Join(c.workdir, c.LocalConfig),
		}
	}

	return c.local.Set(key, value)
}

// SetGlobal sets (or adds) a key only in the per-user (global) config.
func (c *Configs) SetGlobal(key, value string) error {
	if c.global == nil {
		c.global = &Config{
			path: globalConfigFile(),
		}
	}

	return c.global.Set(key, value)
}

// SetEnv sets (or adds) a key in the per-process (env) config. Useful
// for persisting flags during the invocation.
func (c *Configs) SetEnv(key, value string) error {
	if c.env == nil {
		c.env = &Config{
			noWrites: true,
		}
	}

	return c.env.Set(key, value)
}

// UnsetLocal deletes a key from the local config.
func (c *Configs) UnsetLocal(key string) error {
	if c.local == nil {
		return nil
	}

	return c.local.Unset(key)
}

// UnsetGlobal delets a key from the global config.
func (c *Configs) UnsetGlobal(key string) error {
	if c.global == nil {
		return nil
	}

	return c.global.Unset(key)
}

// Keys returns a list of all keys from all available scopes. Every key has section and possibly
// a subsection. They are seprated by dots. The subsection itself may contain dots. The final
// key name and the section MUST NOT contain dots.
//
// Examples
//   - remote.gist.gopass.pw.path -> section: remote, subsection: gist.gopass.pw, key: path
//   - core.timeout -> section: core, key: timeout
func (c *Configs) Keys() []string {
	keys := make([]string, 0, 128)

	for _, cfg := range []*Config{
		c.Preset,
		c.system,
		c.global,
		c.local,
		c.worktree,
		c.env,
	} {
		if cfg == nil {
			continue
		}
		for k := range cfg.vars {
			keys = append(keys, k)
		}
	}

	return set.Sorted(keys)
}

// List returns all keys matching the given prefix. The prefix can be empty,
// then this is identical to Keys().
func (c *Configs) List(prefix string) []string {
	return set.SortedFiltered(c.Keys(), func(k string) bool {
		return strings.HasPrefix(k, prefix)
	})
}

// ListSections returns a sorted list of all sections.
func (c *Configs) ListSections() []string {
	return set.Sorted(set.Apply(c.Keys(), func(k string) string {
		section, _, _ := splitKey(k)

		return section
	}))
}

// ListSubsections returns a sorted list of all subsections
// in the given section.
func (c *Configs) ListSubsections(wantSection string) []string {
	// apply extracts the subsection and matches it to the empty string
	// if it doesn't belong to the section we're looking for. Then the
	// filter func filters out any empty string.
	return set.SortedFiltered(set.Apply(c.Keys(), func(k string) string {
		section, subsection, _ := splitKey(k)
		if section != wantSection {
			return ""
		}

		return subsection
	}), func(s string) bool {
		return s != ""
	})
}
