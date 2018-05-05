package action

import (
	"context"
	"fmt"
	"sort"

	"github.com/justwatchcom/gopass/pkg/out"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

// Config handles changes to the gopass configuration
func (s *Action) Config(ctx context.Context, c *cli.Context) error {
	if len(c.Args()) < 1 {
		s.printConfigValues(ctx, "")
		return nil
	}

	if len(c.Args()) == 1 {
		s.printConfigValues(ctx, "", c.Args()[0])
		return nil
	}

	if len(c.Args()) > 2 {
		return ExitError(ctx, ExitUsage, nil, "Usage: %s config key value", s.Name)
	}

	if err := s.setConfigValue(ctx, c.String("store"), c.Args()[0], c.Args()[1]); err != nil {
		return ExitError(ctx, ExitUnknown, err, "Error setting config value")
	}
	return nil
}

func (s *Action) printConfigValues(ctx context.Context, store string, needles ...string) {
	prefix := ""
	if len(needles) < 1 {
		out.Print(ctx, "root store config:")
		prefix = "  "
	}

	m := s.cfg.Root.ConfigMap()
	if store == "" {
		for _, k := range filterMap(m, needles) {
			out.Print(ctx, "%s%s: %s", prefix, k, m[k])
		}
	}
	for mp, sc := range s.cfg.Mounts {
		if store != "" && mp != store {
			continue
		}
		if len(needles) < 1 {
			out.Print(ctx, "mount '%s' config:", mp)
			mp = "  "
		} else {
			mp += "/"
		}
		sm := sc.ConfigMap()
		for _, k := range filterMap(sm, needles) {
			if sm[k] != m[k] || store != "" {
				out.Print(ctx, "%s%s: %s", mp, k, sm[k])
			}
		}
	}
}

func filterMap(haystack map[string]string, needles []string) []string {
	out := make([]string, 0, len(haystack))
	for k := range haystack {
		if !contains(needles, k) {
			continue
		}
		out = append(out, k)
	}
	sort.Strings(out)
	return out
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

func (s *Action) setConfigValue(ctx context.Context, store, key, value string) error {
	if err := s.cfg.SetConfigValue(store, key, value); err != nil {
		return errors.Wrapf(err, "failed to set config value '%s'", key)
	}
	s.printConfigValues(ctx, store, key)
	return nil
}

// ConfigComplete will print the list of valid config keys
func (s *Action) ConfigComplete(c *cli.Context) {
	cm := s.cfg.Root.ConfigMap()
	keys := make([]string, 0, len(cm))
	for k := range cm {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintln(stdout, k)
	}
}
