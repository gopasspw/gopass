package action

import (
	"context"
	"errors"
	"fmt"

	"github.com/gopasspw/gopass/internal/tree"

	"github.com/gopasspw/gopass/internal/diff"
	"github.com/gopasspw/gopass/internal/notify"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

// Sync all stores with their remotes
func (s *Action) Sync(c *cli.Context) error {
	return s.sync(ctxutil.WithGlobalFlags(c), c.String("store"))
}

func (s *Action) sync(ctx context.Context, store string) error {
	out.Printf(ctx, "🚥 Syncing with all remotes ...")

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
	out.OKf(ctx, "All done")

	// Calculate number of changed entries.
	// This is a rough estimate as additions and deletions
	// might cancel each other out.
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
	out.Printf(ctxno, color.GreenString("[%s] ", name))

	sub, err := s.Store.GetSubStore(mp)
	if err != nil {
		out.Errorf(ctx, "Failed to get sub store %q: %s", name, err)
		return fmt.Errorf("failed to get sub stores (%s)", err)
	}

	if sub == nil {
		out.Errorf(ctx, "Failed to get sub stores '%s: nil'", name)
		return fmt.Errorf("failed to get sub stores (nil)")
	}

	l, err := sub.List(ctx, "")
	if err != nil {
		out.Errorf(ctx, "Failed to list store: %s", err)
	}

	// TODO: Remove this hard coded check
	if sub.Storage().Name() == "fs" {
		out.Printf(ctxno, "\n   WARNING: Mount uses Storage backend 'fs'. Not syncing!\n")
	} else {
		out.Printf(ctxno, "\n   "+color.GreenString("git pull and push ... "))
		if err := sub.Storage().Push(ctx, "", ""); err != nil {
			if errors.Is(err, store.ErrGitNoRemote) {
				out.Printf(ctx, "Skipped (no remote)")
				debug.Log("Failed to push %q to its remote: %s", name, err)
				return err
			}

			out.Errorf(ctx, "Failed to push %q to its remote: %s", name, err)
			return err
		}
		out.Printf(ctxno, color.GreenString("OK"))

		ln, err := sub.List(ctx, "")
		if err != nil {
			out.Errorf(ctx, "Failed to list store: %s", err)
		}

		added, removed := diff.List(l, ln)
		debug.Log("diff - added: %d - removed: %d", added, removed)
		if added > 0 {
			out.Printf(ctxno, color.GreenString(" (Added %d entries)", added))
		}
		if removed > 0 {
			out.Printf(ctxno, color.GreenString(" (Removed %d entries)", removed))
		}
		if added < 1 && removed < 1 {
			out.Printf(ctxno, color.GreenString(" (no changes)"))
		}
	}

	debug.Log("Syncing Mount %s. Exportkeys: %t", mp, ctxutil.IsExportKeys(ctx))
	var exported bool
	if ctxutil.IsExportKeys(ctx) {
		// import keys
		out.Printf(ctxno, "\n   "+color.GreenString("importing missing keys ... "))
		if err := sub.ImportMissingPublicKeys(ctx); err != nil {
			out.Errorf(ctx, "Failed to import missing public keys for %q: %s", name, err)
			return err
		}
		out.Printf(ctxno, color.GreenString("OK"))

		// export keys
		out.Printf(ctxno, "\n   "+color.GreenString("exporting missing keys ... "))
		rs, err := sub.GetRecipients(ctx, "")
		if err != nil {
			out.Errorf(ctx, "Failed to load recipients for %q: %s", name, err)
			return err
		}
		exported, err = sub.ExportMissingPublicKeys(ctx, rs)
		if err != nil {
			out.Errorf(ctx, "Failed to export missing public keys for %q: %s", name, err)
			return err
		}
	}

	// only run second push if we did export any keys
	if exported {
		if err := sub.Storage().Push(ctx, "", ""); err != nil {
			out.Errorf(ctx, "Failed to push %q to its remote: %s", name, err)
			return err
		}
		out.Printf(ctxno, color.GreenString("OK"))
	} else {
		out.Printf(ctxno, color.GreenString("nothing to do"))
	}
	out.Printf(ctx, "\n   "+color.GreenString("done"))
	return nil
}
