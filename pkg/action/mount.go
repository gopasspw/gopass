package action

import (
	"context"
	"fmt"
	"sort"

	"github.com/gopasspw/gopass/pkg/config"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/pkg/store"
	"github.com/gopasspw/gopass/pkg/store/root"
	"github.com/gopasspw/gopass/pkg/tree/simple"
	"github.com/pkg/errors"

	"github.com/fatih/color"
	"github.com/urfave/cli"
)

// MountRemove removes an existing mount
func (s *Action) MountRemove(ctx context.Context, c *cli.Context) error {
	if len(c.Args()) != 1 {
		return ExitError(ctx, ExitUsage, nil, "Usage: %s mount remove [alias]", s.Name)
	}

	if err := s.Store.RemoveMount(ctx, c.Args()[0]); err != nil {
		out.Error(ctx, "Failed to remove mount: %s", err)
	}

	if err := s.cfg.Save(); err != nil {
		return ExitError(ctx, ExitConfig, err, "failed to write config: %s", err)
	}

	out.Green(ctx, "Password Store %s umounted", c.Args()[0])
	return nil
}

// MountsPrint prints all existing mounts
func (s *Action) MountsPrint(ctx context.Context, c *cli.Context) error {
	if len(s.Store.Mounts()) < 1 {
		out.Cyan(ctx, "No mounts")
		return nil
	}

	root := simple.New(color.GreenString(fmt.Sprintf("gopass (%s)", s.Store.Path())))
	mounts := s.Store.Mounts()
	mps := s.Store.MountPoints()
	sort.Sort(store.ByPathLen(mps))
	for _, alias := range mps {
		path := mounts[alias]
		if err := root.AddMount(alias, path); err != nil {
			out.Error(ctx, "Failed to add mount to tree: %s", err)
		}
	}

	fmt.Fprintln(stdout, root.Format(0))
	return nil
}

// MountsComplete will print a list of existings mount points for bash
// completion
func (s *Action) MountsComplete(*cli.Context) {
	for alias := range s.Store.Mounts() {
		fmt.Fprintln(stdout, alias)
	}
}

// MountAdd adds a new mount
func (s *Action) MountAdd(ctx context.Context, c *cli.Context) error {
	alias := c.Args().Get(0)
	localPath := c.Args().Get(1)
	if alias == "" {
		return ExitError(ctx, ExitUsage, nil, "usage: %s mounts add <alias> [local path]", s.Name)
	}

	if localPath == "" {
		localPath = config.PwStoreDir(alias)
	}

	keys := make([]string, 0, 1)
	if k := c.String("init"); k != "" {
		keys = append(keys, k)
	}

	if s.Store.Exists(ctx, alias) {
		out.Yellow(ctx, "WARNING: shadowing %s entry", alias)
	}

	if err := s.Store.AddMount(ctx, alias, localPath, keys...); err != nil {
		switch e := errors.Cause(err).(type) {
		case root.AlreadyMountedError:
			out.Print(ctx, "Store is already mounted")
			return nil
		case root.NotInitializedError:
			out.Print(ctx, "Mount %s is not yet initialized. Initializing ...", e.Alias())
			if err := s.init(ctx, e.Alias(), e.Path()); err != nil {
				return ExitError(ctx, ExitUnknown, err, "failed to add mount '%s': failed to initialize store: %s", e.Alias(), err)
			}
		default:
			return ExitError(ctx, ExitMount, err, "failed to add mount '%s' to '%s': %s", alias, localPath, err)
		}
	}

	if err := s.cfg.Save(); err != nil {
		return ExitError(ctx, ExitConfig, err, "failed to save config: %s", err)
	}

	out.Green(ctx, "Mounted %s as %s", alias, localPath)
	return nil
}
