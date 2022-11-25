package config

import (
	"fmt"

	"github.com/gopasspw/gopass/internal/config/legacy"
	"github.com/gopasspw/gopass/pkg/debug"
)

func migrateConfigs() error {
	cfg := legacy.LoadWithOptions(true, false)
	if cfg == nil {
		debug.Log("no legacy config found. not migrating.")

		return nil
	}

	c := newGitconfig().LoadAll(cfg.Path)

	for k, v := range cfg.ConfigMap() {
		var fk string
		switch k {
		case "keychain":
			fk = "age.usekeychain"
		case "path":
			fk = "mounts.path"
		case "safecontent":
			fk = "core.showsafecontent"
		default:
			fk = "core." + k
		}

		if err := c.SetGlobal(fk, v); err != nil {
			return fmt.Errorf("failed to write new config: %w", err)
		}
	}
	for alias, path := range cfg.Mounts {
		if err := c.SetGlobal(mpk(alias), path); err != nil {
			return fmt.Errorf("failed to write new config: %w", err)
		}
	}

	debug.Log("migrated legacy config from %s", cfg.ConfigPath)

	return nil
}
