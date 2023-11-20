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

// String implements fmt.Stringer.
func (cs *Configs) String() string {
	return fmt.Sprintf("GitConfigs{Env: %s - System: %s - Global: %s - Local: %s - Worktree: %s}", cs.EnvPrefix, cs.SystemConfig, cs.GlobalConfig, cs.LocalConfig, cs.WorktreeConfig)
}

// LoadAll tries to load all known config files. Missing or invalid files are
// silently ignored. It never fails. The workdir is optional. If non-empty
// this method will try to load a local config from this location.
func (cs *Configs) LoadAll(workdir string) *Configs {
	cs.workdir = workdir

	debug.Log("Loading gitconfigs for %s ...", cs)

	// load the system config, if any
	if os.Getenv(cs.EnvPrefix+"_NOSYSTEM") == "" {
		c, err := LoadConfig(cs.SystemConfig)
		if err != nil {
			debug.Log("[%s] failed to load system config: %s", cs.EnvPrefix, err)
		} else {
			debug.Log("[%s] loaded system config from %s", cs.EnvPrefix, cs.SystemConfig)
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
			debug.Log("[%s] failed to load local config from %s: %s", cs.EnvPrefix, localConfigPath, err)
			// set the path just in case we want to modify / write to it later
			cs.local.path = localConfigPath
		} else {
			debug.Log("[%s] loaded local config from %s", cs.EnvPrefix, localConfigPath)
			cs.local = c
		}
	}
	cs.local.noWrites = cs.NoWrites

	// load the worktree config, if any
	if workdir != "" {
		worktreeConfigPath := filepath.Join(workdir, cs.WorktreeConfig)
		c, err := LoadConfig(worktreeConfigPath)
		if err != nil {
			debug.Log("[%s] failed to load worktree config from %s: %s", cs.EnvPrefix, worktreeConfigPath, err)
			// set the path just in case we want to modify / write to it later
			cs.worktree.path = worktreeConfigPath
		} else {
			debug.Log("[%s] loaded local config from %s", cs.EnvPrefix, worktreeConfigPath)
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
// it's easier to handle this in its own method.
func (cs *Configs) loadGlobalConfigs() string {
	locs := []string{
		globalConfigFile(),
	}

	if cs.GlobalConfig != "" {
		// ~/.gitconfig
		locs = append(locs, filepath.Join(appdir.UserHome(), cs.GlobalConfig))
	}

	for _, p := range locs {
		// GlobalConfig might be set to an empty string to disable it
		// and instead of the XDG_CONFIG_HOME path only.
		if p == "" {
			continue
		}
		cfg, err := LoadConfig(p)
		if err != nil {
			debug.Log("[%s] failed to load global config from %s", cs.EnvPrefix, p)

			continue
		}

		debug.Log("[%s] loaded global config from %s", cs.EnvPrefix, p)
		cs.global = cfg

		return p
	}

	debug.Log("[%s] no global config found", cs.EnvPrefix)

	// set the path in case we want to write to it (create it) later
	cs.global = &Config{
		path: globalConfigFile(),
	}

	return ""
}

// HasGlobalConfig indicates if a per-user config can be found.
func (cs *Configs) HasGlobalConfig() bool {
	return cs.loadGlobalConfigs() != ""
}

// Get returns the value for the given key from the first location that is found.
// Lookup order: env, worktree, local, global, system and presets.
func (cs *Configs) Get(key string) string {
	for _, cfg := range []*Config{
		cs.env,
		cs.worktree,
		cs.local,
		cs.global,
		cs.system,
		cs.Preset,
	} {
		if cfg == nil || cfg.vars == nil {
			continue
		}
		if v, found := cfg.Get(key); found {
			return v
		}
	}

	debug.Log("[%s] no value for %s found", cs.EnvPrefix, key)

	return ""
}

// GetAll returns all values for the given key from the first location that is found.
// See the description of Get for more details.
func (cs *Configs) GetAll(key string) []string {
	for _, cfg := range []*Config{
		cs.env,
		cs.worktree,
		cs.local,
		cs.global,
		cs.system,
		cs.Preset,
	} {
		if cfg == nil || cfg.vars == nil {
			continue
		}
		if vs, found := cfg.GetAll(key); found {
			return vs
		}
	}

	debug.Log("[%s] no value for %s found", cs.EnvPrefix, key)

	return nil
}

// GetGlobal specifically ask the per-user (global) config for a key.
func (cs *Configs) GetGlobal(key string) string {
	if cs.global == nil {
		return ""
	}

	if v, found := cs.global.Get(key); found {
		return v
	}

	debug.Log("[%s] no value for %s found", cs.EnvPrefix, key)

	return ""
}

// GetLocal specifically asks the per-directory (local) config for a key.
func (cs *Configs) GetLocal(key string) string {
	if cs.local == nil {
		return ""
	}

	if v, found := cs.local.Get(key); found {
		return v
	}

	debug.Log("[%s] no value for %s found", cs.EnvPrefix, key)

	return ""
}

// IsSet returns true if this key is set in any of our configs.
func (cs *Configs) IsSet(key string) bool {
	for _, cfg := range []*Config{
		cs.env,
		cs.worktree,
		cs.local,
		cs.global,
		cs.system,
		cs.Preset,
	} {
		if cfg != nil && cfg.IsSet(key) {
			return true
		}
	}

	return false
}

// SetLocal sets (or adds) a key only in the per-directory (local) config.
func (cs *Configs) SetLocal(key, value string) error {
	if cs.local == nil {
		if cs.workdir == "" {
			return fmt.Errorf("no workdir set")
		}
		cs.local = &Config{
			path: filepath.Join(cs.workdir, cs.LocalConfig),
		}
	}

	return cs.local.Set(key, value)
}

// SetGlobal sets (or adds) a key only in the per-user (global) config.
func (cs *Configs) SetGlobal(key, value string) error {
	if cs.global == nil {
		cs.global = &Config{
			path: globalConfigFile(),
		}
	}

	return cs.global.Set(key, value)
}

// SetEnv sets (or adds) a key in the per-process (env) config. Useful
// for persisting flags during the invocation.
func (cs *Configs) SetEnv(key, value string) error {
	if cs.env == nil {
		cs.env = &Config{
			noWrites: true,
		}
	}

	return cs.env.Set(key, value)
}

// UnsetLocal deletes a key from the local config.
func (cs *Configs) UnsetLocal(key string) error {
	if cs.local == nil {
		return nil
	}

	return cs.local.Unset(key)
}

// UnsetGlobal deletes a key from the global config.
func (cs *Configs) UnsetGlobal(key string) error {
	if cs.global == nil {
		return nil
	}

	return cs.global.Unset(key)
}

// Keys returns a list of all keys from all available scopes. Every key has section and possibly
// a subsection. They are separated by dots. The subsection itself may contain dots. The final
// key name and the section MUST NOT contain dots.
//
// Examples
//   - remote.gist.gopass.pw.path -> section: remote, subsection: gist.gopass.pw, key: path
//   - core.timeout -> section: core, key: timeout
func (cs *Configs) Keys() []string {
	keys := make([]string, 0, 128)

	for _, cfg := range []*Config{
		cs.Preset,
		cs.system,
		cs.global,
		cs.local,
		cs.worktree,
		cs.env,
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
func (cs *Configs) List(prefix string) []string {
	return set.SortedFiltered(cs.Keys(), func(k string) bool {
		return strings.HasPrefix(k, prefix)
	})
}

// ListSections returns a sorted list of all sections.
func (cs *Configs) ListSections() []string {
	return set.Sorted(set.Apply(cs.Keys(), func(k string) string {
		section, _, _ := splitKey(k)

		return section
	}))
}

// ListSubsections returns a sorted list of all subsections
// in the given section.
func (cs *Configs) ListSubsections(wantSection string) []string {
	// apply extracts the subsection and matches it to the empty string
	// if it doesn't belong to the section we're looking for. Then the
	// filter func filters out any empty string.
	return set.SortedFiltered(set.Apply(cs.Keys(), func(k string) string {
		section, subsection, _ := splitKey(k)
		if section != wantSection {
			return ""
		}

		return subsection
	}), func(s string) bool {
		return s != ""
	})
}
