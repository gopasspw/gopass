package action

import (
	"context"
	"errors"
	"fmt"

	"github.com/gopasspw/gopass/internal/action/exit"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/set"
	istore "github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/urfave/cli/v2"
)

// Config handles changes to the gopass configuration.
func (s *Action) Config(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	store := c.String("store")
	if c.Args().Len() < 1 {
		s.printConfigValues(ctx, store)

		return nil
	}

	if c.Args().Len() == 1 {
		s.printConfigValues(ctx, store, c.Args().Get(0))

		return nil
	}

	if c.Args().Len() > 2 {
		return exit.Error(exit.Usage, nil, "Usage: %s config key value", s.Name)
	}

	// need to initialize sub-stores so we can update their configs
	if err := s.IsInitialized(c); err != nil {
		return err
	}

	if err := s.setConfigValue(ctx, store, c.Args().Get(0), c.Args().Get(1)); err != nil {
		return exit.Error(exit.Unknown, err, "Error setting config value: %s", err)
	}

	return nil
}

func (s *Action) printConfigValues(ctx context.Context, store string, needles ...string) {
	for _, k := range set.SortedFiltered(s.cfg.Keys(store), func(e string) bool {
		return contains(needles, e)
	}) {
		v := s.cfg.GetM(store, k)
		// if only a single key is requested, print only the value
		// useful for scriping, e.g. `$ cd $(gopass config path)`.
		if len(needles) == 1 {
			out.Printf(ctx, "%s", v)

			continue
		}
		out.Printf(ctx, "%s = %s", k, v)
	}
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
	debug.Log("setting %s to %s for %q", key, value, store)

	if err := s.cfg.Set(store, key, value); err != nil {
		return fmt.Errorf("failed to set config value %q: %w", key, err)
	}

	st := s.Store.Storage(ctx, store)
	if st == nil {
		return fmt.Errorf("storage not available")
	}

	if err := st.Add(ctx, "config"); err != nil && !errors.Is(err, istore.ErrGitNotInit) {
		return fmt.Errorf("failed to stage config file: %w", err)
	}
	if err := st.Commit(ctx, "Update config"); err != nil && !errors.Is(err, istore.ErrGitNotInit) {
		return fmt.Errorf("failed to commit config: %w", err)
	}

	s.printConfigValues(ctx, store, key)

	return nil
}

func (s *Action) configKeys() []string {
	return s.cfg.Keys("")
}

// ConfigComplete will print the list of valid config keys.
func (s *Action) ConfigComplete(c *cli.Context) {
	for _, k := range s.configKeys() {
		fmt.Fprintln(stdout, k)
	}
}
