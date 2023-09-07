package config

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/gitconfig"
)

const (
	DefaultLength     = 24
	DefaultXKCDLength = 4
)

var (
	envPrefix    = "GOPASS_CONFIG"
	systemConfig = "/etc/gopass/config"
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

// New initializes a new gopass config. It will handle legacy configs as well.
func New() *Config {
	return newWithOptions(false)
}

// NewNoWrites initializes a new config that does not allow writes. For use in tests.
func NewNoWrites() *Config {
	return newWithOptions(true)
}

func newWithOptions(noWrites bool) *Config {
	c := &Config{
		cfgs: make(map[string]*gitconfig.Configs, 42),
	}

	// if there is no per-user gitconfig we try to migrate
	// an existing config. But we will leave it around for
	// gopass fsck to (optionaly) clean it up.
	if nm := os.Getenv("GOPASS_CONFIG_NO_MIGRATE"); !HasGlobalConfig() && nm == "" {
		if err := migrateConfigs(); err != nil {
			debug.Log("failed to migrate: %s", err)
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

// GetM returns the given key from the mount or the root config if mount is empty.
func (c *Config) GetM(mount, key string) string {
	if mount == "" {
		return c.root.Get(key)
	}

	if cfg := c.cfgs[mount]; cfg != nil {
		return cfg.Get(key)
	}

	return ""
}

// GetBool returns true if the value of the key evaluates to "true".
// Otherwise it returns false.
func (c *Config) GetBool(key string) bool {
	if strings.ToLower(strings.TrimSpace(c.Get(key))) == "true" {
		return true
	}

	return false
}

// GetInt returns the integer value of the key if it can be parsed.
// Otherwise it returns 0.
func (c *Config) GetInt(key string) int {
	iv, err := strconv.Atoi(c.Get(key))
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
	if mount == "" {
		return c.root.SetGlobal(key, value)
	}

	if mount == "<root>" {
		return c.root.SetLocal(key, value)
	}

	if cfg := c.cfgs[mount]; cfg != nil {
		return cfg.SetLocal(key, value)
	}

	return nil
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

// SetPath is a short cut to set the root store path.
func (c *Config) SetPath(path string) error {
	return c.Set("", "mounts.path", path)
}

// SetMountPath is a short cut to set a mount to a path.
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
	if mount == "" {
		return c.root.Keys()
	}

	if cfg := c.cfgs[mount]; cfg != nil {
		return cfg.Keys()
	}

	return nil
}

// DefaultLengthFromEnv will determine the password length from the env variable
// GOPASS_PW_DEFAULT_LENGTH or fallback to the hard-coded default length.
// If the env variable is set by the user and is valid, the boolean return value
// will be true, otherwise it will be false.
func DefaultLengthFromEnv(ctx context.Context) (int, bool) {
	def := DefaultLength
	cfg := FromContext(ctx)

	if l := cfg.GetInt("generate.length"); l > 0 {
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
