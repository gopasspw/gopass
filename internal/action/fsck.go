package action

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/gopasspw/gopass/internal/action/exit"
	"github.com/gopasspw/gopass/internal/config/legacy"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store/leaf"
	"github.com/gopasspw/gopass/internal/store/root"
	"github.com/gopasspw/gopass/internal/tree"
	"github.com/gopasspw/gopass/pkg/appdir"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/fsutil"
	"github.com/gopasspw/gopass/pkg/termio"
	"github.com/urfave/cli/v3"
)

// Fsck checks the store integrity.
func (s *auditHandler) Fsck(ctx context.Context, cmd *cli.Command) error {
	_ = s.rem.Reset("fsck")

	store := cmd.String("store")
	filter := cmd.Args().First()

	ctx = ctxutil.WithGlobalFlags(ctx, cmd)
	if cmd.IsSet("decrypt") {
		ctx = leaf.WithFsckDecrypt(ctx, cmd.Bool("decrypt"))
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
	pwList, err := s.fsckEntries(ctx, store, filter)
	if err != nil {
		return err
	}

	bar := termio.NewProgressBar(int64(len(pwList)) + 1)
	bar.Hidden = ctxutil.IsHidden(ctx)
	progress := ctxutil.ProgressCallback(func() {
		bar.Inc()
	})
	ctx = out.AddPrefix(ctx, "\n")

	// the main work in done by the sub stores.
	if err := s.Store.Fsck(ctx, store, filter, progress); err != nil {
		return exit.Error(exit.Fsck, err, "fsck found errors: %s", err)
	}
	bar.Done()

	return nil
}

func (s *auditHandler) fsckEntries(ctx context.Context, store, filter string) ([]string, error) {
	if store != "" {
		return s.fsckEntriesForStore(ctx, store, filter)
	}

	t, err := s.Store.Tree(ctx)
	if err != nil {
		return nil, exit.Error(exit.Unknown, err, "failed to list stores: %s", err)
	}

	cleanFilter := strings.Trim(filter, "/")
	if cleanFilter == "" {
		return t.List(tree.INF), nil
	}

	if s.Store.Exists(ctx, cleanFilter) {
		return []string{cleanFilter}, nil
	}

	// We restrict ourselves to the filter.
	t, err = t.FindFolder(cleanFilter)
	if err != nil {
		if _, subErr := s.Store.GetSubStore(cleanFilter); subErr == nil {
			return nil, nil
		}

		return nil, exit.Error(exit.NotFound, nil, "Entry %q not found", filter)
	}

	return t.List(tree.INF), nil
}

func (s *auditHandler) fsckEntriesForStore(ctx context.Context, store, filter string) ([]string, error) {
	subAlias := ""
	if store != "<root>" {
		subAlias = root.CleanMountAlias(store)
		if subAlias == "" {
			return nil, exit.Error(exit.NotFound, nil, "mount %q not found", store)
		}
	}

	sub, err := s.Store.GetSubStore(subAlias)
	if err != nil {
		return nil, exit.Error(exit.NotFound, err, "mount %q not found: %s", store, err)
	}

	if sub == nil {
		return nil, exit.Error(exit.NotInitialized, nil, "store %q is not initialized", store)
	}

	cleanFilter := normalizeStoreFilter(ctx, sub, subAlias, filter)
	if cleanFilter != "" && !sub.Exists(ctx, cleanFilter) && !sub.IsDir(ctx, cleanFilter) {
		return nil, exit.Error(exit.NotFound, nil, "Entry %q not found", filter)
	}

	entries, err := sub.List(ctx, cleanFilter)
	if err != nil {
		return nil, exit.Error(exit.Unknown, err, "failed to list store %q: %s", store, err)
	}

	return entries, nil
}

func normalizeStoreFilter(ctx context.Context, sub *leaf.Store, subAlias, filter string) string {
	cleanFilter := strings.Trim(filter, "/")
	if subAlias == "" || cleanFilter == "" {
		return cleanFilter
	}

	if cleanFilter == subAlias {
		return ""
	}

	if sub.Exists(ctx, cleanFilter) || sub.IsDir(ctx, cleanFilter) {
		return cleanFilter
	}

	if after, ok := strings.CutPrefix(cleanFilter, subAlias+"/"); ok {
		return after
	}

	return cleanFilter
}
