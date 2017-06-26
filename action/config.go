package action

import (
	"fmt"
	"sort"

	"github.com/urfave/cli"
)

// Config handles changes to the gopass configuration
func (s *Action) Config(c *cli.Context) error {
	if len(c.Args()) < 1 {
		return s.printConfigValues()
	}

	if len(c.Args()) == 1 {
		return s.printConfigValues(c.Args()[0])
	}

	if len(c.Args()) > 2 {
		return fmt.Errorf("Usage: gopass config key value")
	}

	return s.setConfigValue(c.Args()[0], c.Args()[1])
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
		return err
	}
	return s.Store.UpdateConfig(cfg)
}
