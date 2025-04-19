package action

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/fatih/color"
	"github.com/gopasspw/gopass/internal/action/exit"
	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/internal/store/root"
	"github.com/gopasspw/gopass/internal/tree"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/set"
	"github.com/urfave/cli/v2"
)

// MountRemove removes an existing mount.
func (s *Action) MountRemove(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	if c.Args().Len() != 1 {
		return exit.Error(exit.Usage, nil, "Usage: %s mount remove [alias]", s.Name)
	}

	if err := s.Store.RemoveMount(ctx, c.Args().Get(0)); err != nil {
		out.Errorf(ctx, "Failed to remove mount: %s", err)
	}

	out.Printf(ctx, "Password Store %s umounted", c.Args().Get(0))

	return nil
}

// MountsPrint prints all existing mounts.
func (s *Action) MountsPrint(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	if len(s.Store.Mounts()) < 1 {
		out.Printf(ctx, "No mounts")

		return nil
	}

	root := tree.New(color.GreenString(fmt.Sprintf("gopass (%s)", s.Store.Path())))
	mounts := s.Store.Mounts()
	mps := s.Store.MountPoints()
	sort.Sort(store.ByPathLen(mps))
	for _, alias := range mps {
		path := mounts[alias]
		if err := root.AddMount(alias, path); err != nil {
			out.Errorf(ctx, "Failed to add mount to tree: %s", err)
		}
	}
	debug.Log("MountsPrint - %+v - %+v", mounts, mps)

	fmt.Fprintln(stdout, root.Format(tree.INF))

	return nil
}

// MountsComplete will print a list of existings mount points for bash
// completion.
func (s *Action) MountsComplete(*cli.Context) {
	for alias := range s.Store.Mounts() {
		fmt.Fprintln(stdout, alias)
	}
}

// MountAdd adds a new mount.
func (s *Action) MountAdd(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	alias := c.Args().Get(0)
	localPath := c.Args().Get(1)
	if alias == "" {
		return exit.Error(exit.Usage, nil, "usage: %s mounts add <alias> [local path]", s.Name)
	}

	if localPath == "" {
		localPath = config.PwStoreDir(alias)
	}

	if s.Store.Exists(ctx, alias) {
		out.Warningf(ctx, "shadowing %s entry", alias)
	}

	if c.Bool("create") && !set.New(alias).IsSubset(set.New(s.Store.MountPoints()...)) {
		debug.Log("creating new mount %s at %s", alias, localPath)

		return s.init(ctx, alias, localPath)
	}

	if err := s.Store.AddMount(ctx, alias, localPath); err != nil {
		var aerr *root.AlreadyMountedError
		if errors.As(err, &aerr) {
			out.Printf(ctx, "Store is already mounted")

			return nil
		}
		var nerr *root.NotInitializedError
		if errors.As(err, &nerr) {
			out.Printf(ctx, "Mount %s is not yet initialized. Please use 'gopass init --store %s' instead", nerr.Alias(), nerr.Alias())

			return nerr
		}

		return exit.Error(exit.Mount, err, "failed to add mount %q to %q: %s", alias, localPath, err)
	}

	out.Printf(ctx, "Mounted %s as %s", alias, localPath)

	return nil
}

// MountsVersions prints the backend versions for each mount.
func (s *Action) MountsVersions(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)

	cryptoVer := versionInfo(ctx, s.Store.Crypto(ctx, ""))
	storageVer := versionInfo(ctx, s.Store.Storage(ctx, ""))

	tpl := "%-10s - %10s - %10s\n"
	fmt.Fprintf(stdout, tpl, "<root>", cryptoVer, storageVer)

	// report all used crypto, sync and fs backends.
	for _, mp := range s.Store.MountPoints() {
		cv := versionInfo(ctx, s.Store.Crypto(ctx, mp))
		sv := versionInfo(ctx, s.Store.Storage(ctx, mp))

		fmt.Fprintf(stdout, tpl, mp, cv, sv)
	}

	fmt.Fprintln(stdout)
	fmt.Fprintf(stdout, "Available Crypto Backends: %s\n", strings.Join(backend.CryptoRegistry.BackendNames(), ", "))
	fmt.Fprintf(stdout, "Available Storage Backends: %s\n", strings.Join(backend.StorageRegistry.BackendNames(), ", "))

	return nil
}

type versioner interface {
	Name() string
	Version(context.Context) semver.Version
}

func versionInfo(ctx context.Context, v versioner) string {
	if v == nil {
		return "<none>"
	}

	return fmt.Sprintf("%s %s", v.Name(), v.Version(ctx))
}
