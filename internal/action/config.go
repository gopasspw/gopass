package action

import (
	"context"
	"fmt"
	"sort"

	"github.com/gopasspw/gopass/internal/action/exit"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/urfave/cli/v2"
)

// Config handles changes to the gopass configuration.
func (s *Action) Config(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	if c.Args().Len() < 1 {
		s.printConfigValues(ctx)
		return nil
	}

	if c.Args().Len() == 1 {
		s.printConfigValues(ctx, c.Args().Get(0))
		return nil
	}

	if c.Args().Len() > 2 {
		return exit.Error(exit.Usage, nil, "Usage: %s config key value", s.Name)
	}

	if err := s.setConfigValue(ctx, c.Args().Get(0), c.Args().Get(1)); err != nil {
		return exit.Error(exit.Unknown, err, "Error setting config value")
	}
	return nil
}

func (s *Action) printConfigValues(ctx context.Context, needles ...string) {
	m := s.cfg.ConfigMap()
	for _, k := range filterMap(m, needles) {
		// if only a single key is requested, print only the value
		// useful for scriping, e.g. `$ cd $(gopass config path)`.
		if len(needles) == 1 {
			out.Printf(ctx, "%s", m[k])
			continue
		}
		out.Printf(ctx, "%s: %s", k, m[k])
	}
	for alias, path := range s.cfg.Mounts {
		if len(needles) < 1 {
			out.Printf(ctx, "mount %q => %q", alias, path)
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

func (s *Action) setConfigValue(ctx context.Context, key, value string) error {
	if err := s.cfg.SetConfigValue(key, value); err != nil {
		return fmt.Errorf("failed to set config value %q: %w", key, err)
	}
	s.printConfigValues(ctx, key)
	return nil
}

func (s *Action) configKeys() []string {
	cm := s.cfg.ConfigMap()
	keys := make([]string, 0, len(cm)+1)
	for k := range cm {
		keys = append(keys, k)
	}
	keys = append(keys, "remote")
	sort.Strings(keys)

	return keys
}

// ConfigComplete will print the list of valid config keys.
func (s *Action) ConfigComplete(c *cli.Context) {
	for _, k := range s.configKeys() {
		fmt.Fprintln(stdout, k)
	}
}
