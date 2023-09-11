package action

import (
	"context"
	"os"
	"path/filepath"

	"github.com/gopasspw/gopass/internal/action/exit"
	"github.com/gopasspw/gopass/internal/config/legacy"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store/leaf"
	"github.com/gopasspw/gopass/internal/tree"
	"github.com/gopasspw/gopass/pkg/appdir"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/fsutil"
	"github.com/gopasspw/gopass/pkg/termio"
	"github.com/urfave/cli/v2"
)

// Fsck checks the store integrity.
func (s *Action) Fsck(c *cli.Context) error {
	_ = s.rem.Reset("fsck")

	filter := c.Args().First()

	ctx := ctxutil.WithGlobalFlags(c)
	if c.IsSet("decrypt") {
		ctx = leaf.WithFsckDecrypt(ctx, c.Bool("decrypt"))
	}

	out.Printf(ctx, "Checking password store integrity ...")

	// clean up any previous config locations.
	for _, oldCfg := range append(legacy.ConfigLocations(), filepath.Join(appdir.UserHome(), ".gopass.yml")) {
		if fsutil.IsFile(oldCfg) {
			if err := os.Remove(oldCfg); err != nil {
				out.Errorf(ctx, "Failed to remove old gopass config %s: %s", oldCfg, err)
			}
		}
	}

	// display progress bar.
	pwList, err := s.fsckEntries(ctx, filter)
	if err != nil {
		return err
	}

	bar := termio.NewProgressBar(int64(len(pwList)) + 1)
	bar.Hidden = ctxutil.IsHidden(ctx)
	ctx = ctxutil.WithProgressCallback(ctx, func() {
		bar.Inc()
	})
	ctx = out.AddPrefix(ctx, "\n")

	// the main work in done by the sub stores.
	if err := s.Store.Fsck(ctx, filter); err != nil {
		return exit.Error(exit.Fsck, err, "fsck found errors: %s", err)
	}
	bar.Done()

	return nil
}

func (s *Action) fsckEntries(ctx context.Context, filter string) ([]string, error) {
	t, err := s.Store.Tree(ctx)
	if err != nil {
		return nil, exit.Error(exit.Unknown, err, "failed to list stores: %s", err)
	}

	if filter == "" {
		return t.List(tree.INF), nil
	}

	if s.Store.Exists(ctx, filter) {
		return []string{filter}, nil
	}

	// We restrict ourselves to the filter.
	t, err = t.FindFolder(filter)
	if err != nil {
		return nil, exit.Error(exit.NotFound, nil, "Entry %q not found", filter)
	}

	return t.List(tree.INF), nil
}
