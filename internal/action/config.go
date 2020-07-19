package action

import (
	"context"
	"fmt"
	"sort"

	"github.com/gopasspw/gopass/internal/backend/storage/ondisk"
	"github.com/gopasspw/gopass/internal/debug"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

// Config handles changes to the gopass configuration
func (s *Action) Config(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	if c.Args().Len() < 1 {
		s.printConfigValues(ctx, "")
		return nil
	}

	if c.Args().Len() == 1 {
		s.printConfigValues(ctx, "", c.Args().Get(0))
		return nil
	}

	if c.Args().Len() > 2 {
		return ExitError(ExitUsage, nil, "Usage: %s config key value", s.Name)
	}

	if err := s.setConfigValue(ctx, c.String("store"), c.Args().Get(0), c.Args().Get(1)); err != nil {
		return ExitError(ExitUnknown, err, "Error setting config value")
	}
	return nil
}

func (s *Action) printConfigValues(ctx context.Context, store string, needles ...string) {
	prefix := ""
	if len(needles) < 1 {
		out.Print(ctx, "root store config:")
		prefix = "  "
	}

	m := s.cfg.ConfigMap()
	if store == "" {
		for _, k := range filterMap(m, needles) {
			out.Print(ctx, "%s%s: %s", prefix, k, m[k])
		}
	}
	for alias, path := range s.cfg.Mounts {
		if store != "" && alias != store {
			continue
		}
		if len(needles) < 1 {
			out.Print(ctx, "mount '%s' => '%s'", alias, path)
		}
		storage, ok := s.getOnDiskStorage(ctx, store)
		if !ok {
			// not ondisk
			continue
		}
		rcfg, err := storage.GetRemote(ctx)
		if err != nil {
			debug.Log("failed to read remote config: %s", err)
			continue
		}
		out.Print(ctx, "  remote: %s", rcfg)
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
	if key == "remote" && value != "" {
		return s.setRemoteConfig(ctx, store, value)
	}
	if err := s.cfg.SetConfigValue(key, value); err != nil {
		return errors.Wrapf(err, "failed to set config value '%s'", key)
	}
	s.printConfigValues(ctx, store, key)
	return nil
}

func (s *Action) getOnDiskStorage(ctx context.Context, store string) (*ondisk.OnDisk, bool) {
	_, sub, err := s.Store.GetSubStore(ctx, store)
	if err != nil {
		debug.Log("failed to get sub store: %s", err)
		return nil, false
	}

	storage, ok := sub.Storage().(*ondisk.OnDisk)
	if !ok {
		return nil, false
	}

	return storage, true
}

func (s *Action) setRemoteConfig(ctx context.Context, store, urlStr string) error {
	storage, ok := s.getOnDiskStorage(ctx, store)
	if !ok {
		debug.Log("setting remote not supported")
		return nil
	}
	return storage.SetRemote(ctx, urlStr)
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

// ConfigComplete will print the list of valid config keys
func (s *Action) ConfigComplete(c *cli.Context) {
	for _, k := range s.configKeys() {
		fmt.Fprintln(stdout, k)
	}
}
