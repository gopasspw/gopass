package action

import (
	"context"
	"fmt"
	"sort"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/config"
	"github.com/justwatchcom/gopass/store"
	"github.com/justwatchcom/gopass/utils/tree/simple"
	"github.com/urfave/cli"
)

// MountRemove removes an existing mount
func (s *Action) MountRemove(ctx context.Context, c *cli.Context) error {
	if len(c.Args()) != 1 {
		return s.exitError(ctx, ExitUsage, nil, "Usage: %s mount remove [alias]", s.Name)
	}

	if err := s.Store.RemoveMount(c.Args()[0]); err != nil {
		fmt.Println(color.RedString("Failed to remove mount: %s", err))
	}

	if err := s.cfg.Save(); err != nil {
		return s.exitError(ctx, ExitConfig, err, "failed to write config: %s", err)
	}

	fmt.Println(color.GreenString("Password Store %s umounted", c.Args()[0]))
	return nil
}

// MountsPrint prints all existing mounts
func (s *Action) MountsPrint(ctx context.Context, c *cli.Context) error {
	if len(s.Store.Mounts()) < 1 {
		fmt.Println(color.CyanString("No mounts"))
		return nil
	}

	root := simple.New(color.GreenString(fmt.Sprintf("gopass (%s)", s.Store.Path())))
	mounts := s.Store.Mounts()
	mps := s.Store.MountPoints()
	sort.Sort(store.ByPathLen(mps))
	for _, alias := range mps {
		path := mounts[alias]
		if err := root.AddMount(alias, path); err != nil {
			fmt.Println(color.RedString("Failed to add mount to tree: %s", err))
		}
	}

	fmt.Println(root.Format(0))
	return nil
}

// MountsComplete will print a list of existings mount points for bash
// completion
func (s *Action) MountsComplete(*cli.Context) {
	for alias := range s.Store.Mounts() {
		fmt.Println(alias)
	}
}

// MountAdd adds a new mount
func (s *Action) MountAdd(ctx context.Context, c *cli.Context) error {
	alias := c.Args().Get(0)
	localPath := c.Args().Get(1)
	if alias == "" {
		return s.exitError(ctx, ExitUsage, nil, "usage: %s mount add <alias> [local path]", s.Name)
	}

	if localPath == "" {
		localPath = config.PwStoreDir(alias)
	}

	keys := make([]string, 0, 1)
	if k := c.String("init"); k != "" {
		keys = append(keys, k)
	}

	if s.Store.Exists(ctx, alias) {
		fmt.Printf(color.YellowString("WARNING: shadowing %s entry\n"), alias)
	}

	if err := s.Store.AddMount(ctx, alias, localPath, keys...); err != nil {
		return s.exitError(ctx, ExitMount, err, "failed to add mount '%s' to '%s': %s", alias, localPath, err)
	}

	if err := s.cfg.Save(); err != nil {
		return s.exitError(ctx, ExitConfig, err, "failed to save config: %s", err)
	}

	fmt.Println(color.GreenString("Mounted %s as %s", alias, localPath))
	return nil
}
