package action

import (
	"fmt"
	"sort"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

// Config handles changes to the gopass configuration
func (s *Action) Config(c *cli.Context) error {
	if len(c.Args()) < 1 {
		if err := s.printConfigValues(); err != nil {
			return s.exitError(ExitUnknown, err, "Error printing config")
		}
		return nil
	}

	if len(c.Args()) == 1 {
		if err := s.printConfigValues(c.Args()[0]); err != nil {
			return s.exitError(ExitUnknown, err, "Error printing config value")
		}
		return nil
	}

	if len(c.Args()) > 2 {
		return s.exitError(ExitUsage, nil, "Usage: %s config key value", s.Name)
	}

	if err := s.setConfigValue(c.Args()[0], c.Args()[1]); err != nil {
		return s.exitError(ExitUnknown, err, "Error setting config value")
	}
	return nil
}

func (s *Action) printConfigValues(filter ...string) error {
	m := s.Store.Config().ConfigMap()
	out := make([]string, 0, len(m))
	for k := range m {
		if !contains(filter, k) {
			continue
		}
		out = append(out, k)
	}
	sort.Strings(out)
	for _, k := range out {
		fmt.Printf("%s: %s\n", k, m[k])
	}
	return nil
}

func contains(haystack []string, needle string) bool {
	if len(haystack) < 1 {
		return true
	}
	for _, blade := range haystack {
		if blade == needle {
			return true
		}
	}
	return false
}

func (s *Action) setConfigValue(key, value string) error {
	cfg := s.Store.Config()
	if err := cfg.SetConfigValue(key, value); err != nil {
		return errors.Wrapf(err, "failed to set config value '%s'", key)
	}
	if err := s.Store.UpdateConfig(cfg); err != nil {
		return errors.Wrapf(err, "failed to update config")
	}
	return s.printConfigValues(key)
}
