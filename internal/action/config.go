package action

import (
	"context"
	"fmt"

	"github.com/gopasspw/gopass/internal/action/exit"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/set"
	"github.com/urfave/cli/v2"
)

// Config handles changes to the gopass configuration.
// It can be used to print the whole config, a single key, or to set a new
// value for a given key.
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

	// sub-stores need to have been initialized so we can update their local configs.
	// special case: we can always update the global config. NB: IsInitialized initializes the store if nil.
	if inited, err := s.Store.IsInitialized(ctx); err != nil || (store != "" && !inited) {
		return exit.Error(exit.Unknown, err, "Store %s seems uninitialized or cannot be initialized", store)
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

	level, err := s.cfg.SetWithLevel(store, key, value)
	if err != nil {
		return fmt.Errorf("failed to set config value %q: %w", key, err)
	}

	// in case of a non-local config change we don't need to track changes in the store
	if level != config.Local && level != config.Worktree {
		debug.Log("did not set local config (was config level %d) in store %q, skipping commit phase", level, store)
		s.printConfigValues(ctx, store, key)

		return nil
	}

	st := s.Store.Storage(ctx, store)
	if st == nil {
		return fmt.Errorf("storage not available")
	}

	configFile := "config"
	if level == config.Worktree {
		configFile = "config.worktree"
	}

	// notice that the cfg.Set above should have created the local config file.
	if !st.Exists(ctx, configFile) {
		return fmt.Errorf("local config file %s didn't exist in store, this is unexpected", configFile)
	}

	if err := st.TryAdd(ctx, configFile); err != nil {
		return fmt.Errorf("failed to Add local config %q: %w", configFile, err)
	}
	debug.Log("Added local config for commit")

	if err := st.TryCommit(ctx, "Update config"); err != nil {
		return fmt.Errorf("failed to commit local config: %w", err)
	}
	debug.Log("Committed local config")

	s.printConfigValues(ctx, store, key)

	return nil
}

func (s *Action) configKeys() []string {
	return s.cfg.Keys("")
}

// ConfigComplete will print the list of valid config keys for bash completion.
func (s *Action) ConfigComplete(c *cli.Context) {
	for _, k := range s.configKeys() {
		fmt.Fprintln(stdout, k)
	}
}
