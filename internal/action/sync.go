package action

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/fatih/color"
	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/diff"
	"github.com/gopasspw/gopass/internal/notify"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/internal/store/leaf"
	"github.com/gopasspw/gopass/internal/tree"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/urfave/cli/v2"
)

var autosyncIntervalDays = 3

func init() {
	sv := os.Getenv("GOPASS_AUTOSYNC_INTERVAL")
	if sv == "" {
		return
	}

	iv, err := strconv.Atoi(sv)
	if err != nil {
		return
	}

	autosyncIntervalDays = iv
}

// Sync all stores with their remotes.
func (s *Action) Sync(c *cli.Context) error {
	return s.sync(ctxutil.WithGlobalFlags(c), c.String("store"))
}

func (s *Action) autoSync(ctx context.Context) error {
	if !ctxutil.IsInteractive(ctx) {
		return nil
	}

	if !ctxutil.IsTerminal(ctx) {
		return nil
	}

	if sv := os.Getenv("GOPASS_NO_AUTOSYNC"); sv != "" {
		return nil
	}

	ls := s.rem.LastSeen("autosync")
	debug.Log("autosync - last seen: %s", ls)
	if time.Since(ls) > time.Duration(autosyncIntervalDays)*24*time.Hour {
		_ = s.rem.Reset("autosync")

		return s.sync(ctx, "")
	}

	return nil
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

	// sync all stores (root and all mounted sub stores).
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
	// This is a rough estimate as additions and deletions.
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

// syncMount syncs a single mount.
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

		return fmt.Errorf("failed to get sub stores (%w)", err)
	}

	if sub == nil {
		out.Errorf(ctx, "Failed to get sub stores '%s: nil'", name)

		return fmt.Errorf("failed to get sub stores (nil)")
	}

	l, err := sub.List(ctx, "")
	if err != nil {
		out.Errorf(ctx, "Failed to list store: %s", err)
	}

	out.Printf(ctxno, "\n   "+color.GreenString("%s pull and push ... ", sub.Storage().Name()))
	err = sub.Storage().Push(ctx, "", "")

	switch {
	case err == nil:
		debug.Log("Push succeeded")
		out.Printf(ctxno, color.GreenString("OK"))
	case errors.Is(err, store.ErrGitNoRemote):
		out.Printf(ctx, "Skipped (no remote)")
		debug.Log("Failed to push %q to its remote: %s", name, err)

		return err
	case errors.Is(err, backend.ErrNotSupported):
		out.Printf(ctxno, "Skipped (not supported)")
	case errors.Is(err, store.ErrGitNotInit):
		out.Printf(ctxno, "Skipped (no Git repo)")
	default: // any other error
		out.Errorf(ctx, "Failed to push %q to its remote: %s", name, err)

		return err
	}

	ln, err := sub.List(ctx, "")
	if err != nil {
		out.Errorf(ctx, "Failed to list store: %s", err)
	}
	syncPrintDiff(ctxno, l, ln)

	debug.Log("Syncing Mount %s. Exportkeys: %t", mp, ctxutil.IsExportKeys(ctx))
	if err := syncImportKeys(ctxno, sub, name); err != nil {
		return err
	}
	if ctxutil.IsExportKeys(ctx) {
		if err := syncExportKeys(ctxno, sub, name); err != nil {
			return err
		}
	}
	out.Printf(ctx, "\n   "+color.GreenString("done"))

	return nil
}

func syncImportKeys(ctx context.Context, sub *leaf.Store, name string) error {
	// import keys.
	out.Printf(ctx, "\n   "+color.GreenString("importing missing keys ... "))
	if err := sub.ImportMissingPublicKeys(ctx); err != nil {
		out.Errorf(ctx, "Failed to import missing public keys for %q: %s", name, err)

		return err
	}
	out.Printf(ctx, color.GreenString("OK"))

	return nil
}

func syncExportKeys(ctx context.Context, sub *leaf.Store, name string) error {
	// export keys.
	out.Printf(ctx, "\n   "+color.GreenString("exporting missing keys ... "))
	rs, err := sub.GetRecipients(ctx, "")
	if err != nil {
		out.Errorf(ctx, "Failed to load recipients for %q: %s", name, err)

		return err
	}
	exported, err := sub.ExportMissingPublicKeys(ctx, rs)
	if err != nil {
		out.Errorf(ctx, "Failed to export missing public keys for %q: %s", name, err)

		return err
	}

	// only run second push if we did export any keys.
	if !exported {
		out.Printf(ctx, color.GreenString("nothing to do"))

		return nil
	}

	if err := sub.Storage().Push(ctx, "", ""); err != nil {
		out.Errorf(ctx, "Failed to push %q to its remote: %s", name, err)

		return err
	}
	out.Printf(ctx, color.GreenString("OK"))

	return nil
}

func syncPrintDiff(ctxno context.Context, l, r []string) {
	added, removed := diff.Stat(l, r)
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
