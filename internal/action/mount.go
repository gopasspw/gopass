package action

import (
	"fmt"
	"sort"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/debug"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/internal/store/root"
	"github.com/gopasspw/gopass/internal/tree"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/pkg/errors"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

// MountRemove removes an existing mount
func (s *Action) MountRemove(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	if c.Args().Len() != 1 {
		return ExitError(ExitUsage, nil, "Usage: %s mount remove [alias]", s.Name)
	}

	if err := s.Store.RemoveMount(ctx, c.Args().Get(0)); err != nil {
		out.Error(ctx, "Failed to remove mount: %s", err)
	}

	if err := s.cfg.Save(); err != nil {
		return ExitError(ExitConfig, err, "failed to write config: %s", err)
	}

	out.Green(ctx, "Password Store %s umounted", c.Args().Get(0))
	return nil
}

// MountsPrint prints all existing mounts
func (s *Action) MountsPrint(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	if len(s.Store.Mounts()) < 1 {
		out.Cyan(ctx, "No mounts")
		return nil
	}

	root := tree.New(color.GreenString(fmt.Sprintf("gopass (%s)", s.Store.Path())))
	mounts := s.Store.Mounts()
	mps := s.Store.MountPoints()
	sort.Sort(store.ByPathLen(mps))
	for _, alias := range mps {
		path := mounts[alias]
		if err := root.AddMount(alias, path); err != nil {
			out.Error(ctx, "Failed to add mount to tree: %s", err)
		}
	}
	debug.Log("MountsPrint - %+v - %+v", mounts, mps)

	fmt.Fprintln(stdout, root.Format(-1))
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
func (s *Action) MountAdd(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	alias := c.Args().Get(0)
	localPath := c.Args().Get(1)
	if alias == "" {
		return ExitError(ExitUsage, nil, "usage: %s mounts add <alias> [local path]", s.Name)
	}

	if localPath == "" {
		localPath = config.PwStoreDir(alias)
	}

	if s.Store.Exists(ctx, alias) {
		out.Yellow(ctx, "WARNING: shadowing %s entry", alias)
	}

	if err := s.Store.AddMount(ctx, alias, localPath); err != nil {
		switch e := errors.Cause(err).(type) {
		case root.AlreadyMountedError:
			out.Print(ctx, "Store is already mounted")
			return nil
		case root.NotInitializedError:
			out.Print(ctx, "Mount %s is not yet initialized. Please use 'gopass init --store %s' instead", e.Alias(), e.Alias())
			return e
		default:
			return ExitError(ExitMount, err, "failed to add mount '%s' to '%s': %s", alias, localPath, err)
		}
	}

	if err := s.cfg.Save(); err != nil {
		return ExitError(ExitConfig, err, "failed to save config: %s", err)
	}

	out.Green(ctx, "Mounted %s as %s", alias, localPath)
	return nil
}
