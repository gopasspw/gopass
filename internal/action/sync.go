package action

import (
	"context"
	"fmt"

	"github.com/gopasspw/gopass/internal/tree"

	"github.com/gopasspw/gopass/internal/debug"
	"github.com/gopasspw/gopass/internal/notify"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/pkg/ctxutil"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

// Sync all stores with their remotes
func (s *Action) Sync(c *cli.Context) error {
	return s.sync(ctxutil.WithGlobalFlags(c), c.String("store"))
}

func (s *Action) sync(ctx context.Context, store string) error {
	out.Green(ctx, "Sync starting ...")

	numEntries := 0
	if l, err := s.Store.Tree(ctx); err == nil {
		numEntries = len(l.List(tree.INF))
	}
	numMPs := 0

	mps := s.Store.MountPoints()
	mps = append([]string{""}, mps...)

	// sync all stores (root and all mounted sub stores)
	for _, mp := range mps {
		if store != "" {
			if store != "root" && mp != store {
				continue
			}
			if store == "root" && mp != "" {
				continue
			}
		}

		numMPs++
		_ = s.syncMount(ctx, mp)
	}
	out.Green(ctx, "All done")

	// calculate number of changes entries
	if l, err := s.Store.Tree(ctx); err == nil {
		numEntries = len(l.List(tree.INF)) - numEntries
	}
	diff := ""
	if numEntries > 0 {
		diff = fmt.Sprintf(" Added %d entries", numEntries)
	} else if numEntries < 0 {
		diff = fmt.Sprintf(" Removed %d entries", -1*numEntries)
	}
	_ = notify.Notify(ctx, "gopass - sync", fmt.Sprintf("Finished. Synced %d remotes.%s", numMPs, diff))

	return nil
}

// syncMount syncs a single mount
func (s *Action) syncMount(ctx context.Context, mp string) error {
	ctxno := out.WithNewline(ctx, false)
	name := mp
	if mp == "" {
		name = "<root>"
	}
	out.Print(ctxno, color.GreenString("[%s] ", name))

	ctx, sub, err := s.Store.GetSubStore(ctx, mp)
	if err != nil {
		out.Error(ctx, "Failed to get sub store '%s': %s", name, err)
		return fmt.Errorf("failed to get sub stores (%s)", err)
	}

	if sub == nil {
		out.Error(ctx, "Failed to get sub stores '%s: nil'", name)
		return fmt.Errorf("failed to get sub stores (nil)")
	}

	numMP := 0
	if l, err := sub.List(ctx, ""); err == nil {
		numMP = len(l)
	}

	// TODO: Remove this hard coded check
	if sub.Storage().Name() == "fs" {
		out.Yellow(ctxno, "\n   WARNING: Mount uses Storage backend 'fs'. Not syncing!\n")
	} else {
		out.Print(ctxno, "\n   "+color.GreenString("git pull and push ... "))
		if err := sub.Storage().Push(ctx, "", ""); err != nil {
			if errors.Cause(err) == store.ErrGitNoRemote {
				out.Yellow(ctx, "Skipped (no remote)")
				debug.Log("Failed to push '%s' to its remote: %s", name, err)
				return err
			}

			out.Error(ctx, "Failed to push '%s' to its remote: %s", name, err)
			return err
		}
		out.Print(ctxno, color.GreenString("OK"))

		if l, err := sub.List(ctx, ""); err == nil {
			diff := len(l) - numMP
			if diff > 0 {
				out.Print(ctxno, color.GreenString(" (Added %d entries)", diff))
			} else if diff < 0 {
				out.Print(ctxno, color.GreenString(" (Removed %d entries)", -1*diff))
			} else {
				out.Print(ctxno, color.GreenString(" (no changes)"))
			}
		}
	}

	debug.Log("Syncing Mount %s. Exportkeys: %t", mp, ctxutil.IsExportKeys(ctx))
	var exported bool
	if ctxutil.IsExportKeys(ctx) {
		// import keys
		out.Print(ctxno, "\n   "+color.GreenString("importing missing keys ... "))
		if err := sub.ImportMissingPublicKeys(ctx); err != nil {
			out.Error(ctx, "Failed to import missing public keys for '%s': %s", name, err)
			return err
		}
		out.Print(ctxno, color.GreenString("OK"))

		// export keys
		out.Print(ctxno, "\n   "+color.GreenString("exporting missing keys ... "))
		rs, err := sub.GetRecipients(ctx, "")
		if err != nil {
			out.Error(ctx, "Failed to load recipients for '%s': %s", name, err)
			return err
		}
		exported, err = sub.ExportMissingPublicKeys(ctx, rs)
		if err != nil {
			out.Error(ctx, "Failed to export missing public keys for '%s': %s", name, err)
			return err
		}
	}

	// only run second push if we did export any keys
	if exported {
		if err := sub.Storage().Push(ctx, "", ""); err != nil {
			out.Error(ctx, "Failed to push '%s' to its remote: %s", name, err)
			return err
		}
		out.Print(ctxno, color.GreenString("OK"))
	} else {
		out.Print(ctxno, color.GreenString("nothing to do"))
	}
	out.Print(ctx, "\n   "+color.GreenString("done"))
	return nil
}
