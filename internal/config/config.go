package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/gitconfig"
)

const (
	DefaultPasswordLength = 24
	DefaultXKCDLength     = 4
)

var (
	envPrefix    = "GOPASS_CONFIG"
	systemConfig = "/etc/gopass/config"
)

type Level int

const (
	None Level = iota
	Env
	Worktree
	Local
	Global
	System
	Preset
)

func newGitconfig() *gitconfig.Configs {
	c := gitconfig.New()
	c.EnvPrefix = envPrefix
	c.GlobalConfig = os.Getenv("GOPASS_CONFIG")
	c.SystemConfig = systemConfig

	return c
}

var defaults = map[string]string{
	"core.autopush":      "true",
	"core.autosync":      "true",
	"core.cliptimeout":   "45",
	"core.exportkeys":    "true",
	"core.notifications": "true",
}

// Config is a gopass config handler.
type Config struct {
	root *gitconfig.Configs
	cfgs map[string]*gitconfig.Configs
}

// migrationOpts is a list of config options that were used by gopass
// and need to be migrated to a new name, it maps old name -> new name
// the keys are used in our documentation test to spot legacy options
// that are still used in our codebase.
var migrationOpts = map[string]string{
	// migration done in v1.15.9
	"core.showsafecontent": "show.safecontent",
	"core.autoclip":        "generate.autoclip",
	"core.showautoclip":    "show.autoclip",
}

// New initializes a new gopass config. It will handle legacy configs as well and legacy option names, migrating
// them to their new location and names on a best effort basis. Any system level config or env variables options are
// not migrated.
func New() *Config {
	c := newWithOptions(false)
	// we only migrate options when we are allowed to write them
	c.migrateOptions(migrationOpts)

	return c
}

// NewNoWrites initializes a new config that does not allow writes. For use in tests.
// This does not migrate legacy option names to their correct config section.
func NewNoWrites() *Config {
	return newWithOptions(true)
}

func newWithOptions(noWrites bool) *Config {
	c := &Config{
		cfgs: make(map[string]*gitconfig.Configs, 42),
	}

	// if there is no per-user gitconfig we try to migrate
	// an existing config. But we will leave it around for
	// gopass fsck to (optionally) clean it up.
	if nm := os.Getenv("GOPASS_CONFIG_NO_MIGRATE"); !HasGlobalConfig() && nm == "" {
		if err := migrateConfigs(); err != nil {
			debug.Log("failed to migrate from old config: %s", err)
		}
	}

	// load the global config to get the root path
	c.root = newGitconfig().LoadAll("")
	c.root.NoWrites = noWrites

	rootPath := c.root.Get("mounts.path")
	if rootPath == "" {
		if err := c.SetPath(PwStoreDir("")); err != nil {
			debug.Log("failed to set path: %s", err)
		}
	}
	// load again, this might add a per-store config from the root store
	c.root.LoadAll(rootPath)
	c.root.NoWrites = noWrites

	if rootPath := c.root.Get("mounts.path"); rootPath == "" {
		if err := c.SetPath(PwStoreDir("")); err != nil {
			debug.Log("failed to set path: %s", err)
		}
	}

	// set global defaults
	c.root.Preset = gitconfig.NewFromMap(defaults)

	for _, m := range c.Mounts() {
		c.cfgs[m] = newGitconfig().LoadAll(c.MountPath(m))
		c.cfgs[m].NoWrites = noWrites
	}

	return c
}

// HasGlobalConfig returns true if there is an existing global config.
func HasGlobalConfig() bool {
	return newGitconfig().HasGlobalConfig()
}

// IsSet returns true if the key is set in the root config.
func (c *Config) IsSet(key string) bool {
	return c.root.IsSet(key)
}

// Get returns the given key from the root config.
func (c *Config) Get(key string) string {
	return c.root.Get(key)
}

// GetAll returns all values for the given key.
func (c *Config) GetAll(key string) []string {
	return c.root.GetAll(key)
}

// GetGlobal returns the given key from the root global config.
// This is typically used to prevent a local config override of sensitive config items, e.g. used for integrity checks.
func (c *Config) GetGlobal(key string) string {
	return c.root.GetGlobal(key)
}

// GetM returns the given key from the mount or the root config if mount is empty.
func (c *Config) GetM(mount, key string) string {
	if mount == "" || mount == "<root>" {
		return c.root.Get(key)
	}

	if cfg := c.cfgs[mount]; cfg != nil {
		return cfg.Get(key)
	}

	return ""
}

// GetBool returns true if the value of the key evaluates to "true".
// Otherwise, it returns false.
func (c *Config) GetBool(key string) bool {
	return c.GetBoolM("", key)
}

// GetBoolM returns true if the value of the key evaluates to "true" for the provided mount,
// or the root config if mount is empty.
// Otherwise, it returns false.
func (c *Config) GetBoolM(mount, key string) bool {
	if strings.ToLower(strings.TrimSpace(c.GetM(mount, key))) == "true" {
		return true
	}

	return false
}

// GetInt returns the integer value of the key if it can be parsed.
// Otherwise, it returns 0.
func (c *Config) GetInt(key string) int {
	return c.GetIntM("", key)
}

// GetIntM returns the integer value of the key if it can be parsed for the provided mount,
// or the root config if mount is empty
// Otherwise, it returns 0.
func (c *Config) GetIntM(mount, key string) int {
	iv, err := strconv.Atoi(c.GetM(mount, key))
	if err != nil {
		return 0
	}

	return iv
}

// Set tries to set the key to the given value.
// The mount option is necessary to discern between
// the per-user (global) and possible per-directory (local)
// config files.
//
//   - If mount is empty the setting will be written to the per-user config (global)
//   - If mount has the special value "<root>" the setting will be written to the per-directory config of the root store (local)
//   - If mount has any other value we will attempt to write the setting to the per-directory config of this mount.
//   - If the mount point does not exist we will return nil.
func (c *Config) Set(mount, key, value string) error {
	_, err := c.SetWithLevel(mount, key, value)

	return err
}

// SetWithLevel is the same as Set, but it also returns the level at which the config was set.
// It currently only supports global and local configs.
func (c *Config) SetWithLevel(mount, key, value string) (Level, error) {
	if mount == "" {
		return Global, c.root.SetGlobal(key, value)
	}

	if mount == "<root>" {
		return Local, c.root.SetLocal(key, value)
	}

	if cfg, ok := c.cfgs[mount]; !ok {
		return None, fmt.Errorf("substore %q is not initialized or doesn't exist", mount)
	} else if cfg != nil {
		return Local, cfg.SetLocal(key, value)
	}

	return None, nil
}

// SetEnv overrides a key in the non-persistent layer.
func (c *Config) SetEnv(key, value string) error {
	return c.root.SetEnv(key, value)
}

// Path returns the root store path.
func (c *Config) Path() string {
	return c.Get("mounts.path")
}

// MountPath returns the mount store path.
func (c *Config) MountPath(mountPoint string) string {
	return c.Get(mpk(mountPoint))
}

// SetPath is a shortcut to set the root store path.
func (c *Config) SetPath(path string) error {
	return c.Set("", "mounts.path", path)
}

// SetMountPath is a shortcut to set a mount to a path.
func (c *Config) SetMountPath(mount, path string) error {
	return c.Set("", mpk(mount), path)
}

// mpk for mountPathKey.
func mpk(mount string) string {
	return fmt.Sprintf("mounts.%s.path", mount)
}

// Mounts returns all mount points from the root config.
// Note: Any mounts in local configs are ignored.
func (c *Config) Mounts() []string {
	return c.root.ListSubsections("mounts")
}

// Unset deletes the key from the given config.
func (c *Config) Unset(mount, key string) error {
	if mount == "" {
		return c.root.UnsetGlobal(key)
	}

	if mount == "<root>" {
		return c.root.UnsetLocal(key)
	}

	if cfg := c.cfgs[mount]; cfg != nil {
		return cfg.UnsetLocal(key)
	}

	return nil
}

// Keys returns all keys in the given config.
func (c *Config) Keys(mount string) []string {
	if mount == "" || mount == "<root>" {
		return c.root.Keys()
	}

	if cfg := c.cfgs[mount]; cfg != nil {
		return cfg.Keys()
	}

	return nil
}

// migrateOptions is a best effort migration tool for when we introduce new options. It does not necessarily
// handle worktree and env level options very well.
func (c *Config) migrateOptions(migrations map[string]string) {
	if nm := os.Getenv("GOPASS_CONFIG_NO_MIGRATE"); nm != "" {
		return
	}
	var errs []error
	debug.Log("migrateOptions running")
	for oldK, newK := range migrations {
		found := false
		if val := c.root.GetGlobal(oldK); val != "" {
			debug.Log("migrating option in root global store: %s -> %s ", oldK, newK)
			errs = append(errs, c.root.SetGlobal(newK, val))
			errs = append(errs, c.root.UnsetGlobal(oldK))
			found = true
		}
		if val := c.root.GetLocal(oldK); val != "" {
			debug.Log("migrating option in <root> local store: %s -> %s ", oldK, newK)
			errs = append(errs, c.root.SetLocal(newK, val))
			errs = append(errs, c.root.UnsetLocal(oldK))
			found = true
		}
		for _, m := range c.Mounts() {
			if cfg := c.cfgs[m]; cfg != nil {
				if val := cfg.GetLocal(oldK); val != "" {
					debug.Log("migrating option in local store %s: %s -> %s ", m, oldK, newK)
					errs = append(errs, cfg.SetLocal(newK, val))
					errs = append(errs, cfg.UnsetLocal(oldK))
					found = true
				}
				if val := cfg.Get(oldK); !found && val != "" {
					debug.Log("Found old option %s = %s in config, probably at the worktree or env level, "+
						"or maybe at the system level cannot migrate it.", oldK, val)
				}
			}
		}
	}
	if err := errors.Join(errs...); err != nil {
		debug.Log("Errors encountered while migrating old options: {%v}", err)
	}
}

// DefaultPasswordLengthFromEnv will determine the password length from the env variable
// GOPASS_PW_DEFAULT_LENGTH or fallback to the hard-coded default length.
// If the env variable is set by the user and is valid, the boolean return value
// will be true, otherwise it will be false.
func DefaultPasswordLengthFromEnv(ctx context.Context) (int, bool) {
	def := DefaultPasswordLength
	cfg, mp := FromContext(ctx)

	if l := cfg.GetIntM(mp, "generate.length"); l > 0 {
		def = l
	}

	lengthStr, isSet := os.LookupEnv("GOPASS_PW_DEFAULT_LENGTH")
	if !isSet {
		return def, false
	}
	length, err := strconv.Atoi(lengthStr)
	if err != nil || length < 1 {
		return def, false
	}

	return length, true
}
