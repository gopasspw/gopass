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
	"github.com/gopasspw/gopass/internal/config"
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

var (
	autosyncIntervalDays = 3
	autosyncLastRun      time.Time
)

func init() {
	sv := os.Getenv("GOPASS_AUTOSYNC_INTERVAL")
	if sv == "" {
		return
	}

	debug.Log("GOPASS_AUTOSYNC_INTERVAL is deprecated. Please use autosync.interval")

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
		out.Warning(ctx, "GOPASS_NO_AUTOSYNC is deprecated. Please set core.autosync = false.")

		return nil
	}

	if !config.Bool(ctx, "core.autosync") {
		return nil
	}

	ls := s.rem.LastSeen("autosync")
	debug.Log("autosync - last seen: %s", ls)
	syncInterval := autosyncIntervalDays

	if s.cfg.IsSet("autosync.interval") {
		syncInterval = s.cfg.GetInt("autosync.interval")
	}

	if time.Since(ls) > time.Duration(syncInterval)*24*time.Hour {
		err := s.sync(ctx, "")
		if err != nil {
			autosyncLastRun = time.Now()
		}

		return err
	}

	return nil
}

func (s *Action) sync(ctx context.Context, store string) error {
	// we just did a full sync, no need to run it again
	if time.Since(autosyncLastRun) < 10*time.Second {
		debug.Log("skipping sync. last sync %ds ago", time.Since(autosyncLastRun))

		return nil
	}

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
			if store != "<root>" && mp != store {
				continue
			}
			if store == "<root>" && mp != "" {
				continue
			}
		}

		numMPs++
		_ = s.syncMount(ctx, mp)
	}
	out.OKf(ctx, "All done")

	// If we just sync'ed all stores we can reset the auto-sync interval
	if store == "" {
		_ = s.rem.Reset("autosync")
	}

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

	if numEntries != 0 {
		ctx = config.WithMount(ctx, store)
		_ = notify.Notify(ctx, "gopass - sync", fmt.Sprintf("Finished. Synced %d remotes.%s", numMPs, diff))
	}

	return nil
}

// syncMount syncs a single mount.
func (s *Action) syncMount(ctx context.Context, mp string) error {
	// using GetM here to get the value for this mount, it might be different
	// than the global value
	if as := s.cfg.GetM(mp, "core.autosync"); as == "false" {
		debug.Log("not syncing %s, autosync is disabled for this mount", mp)

		return nil
	}

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

	exportKeys := s.cfg.GetBool("core.exportkeys")
	debug.Log("Syncing Mount %s. Exportkeys: %t", mp, exportKeys)
	if err := syncImportKeys(ctxno, sub, name); err != nil {
		return err
	}
	if exportKeys {
		if err := syncExportKeys(ctxno, sub, name); err != nil {
			return err
		}
	}
	out.Printf(ctx, "\n   "+color.GreenString("done"))

	return nil
}

func syncImportKeys(ctx context.Context, sub *leaf.Store, name string) error {
	// import keys.
	if err := sub.ImportMissingPublicKeys(ctx); err != nil {
		out.Errorf(ctx, "Failed to import missing public keys for %q: %s", name, err)

		return err
	}

	return nil
}

func syncExportKeys(ctx context.Context, sub *leaf.Store, name string) error {
	// export keys.
	rs, err := sub.GetRecipients(ctx, "")
	if err != nil {
		out.Errorf(ctx, "Failed to load recipients for %q: %s", name, err)

		return err
	}
	exported, err := sub.UpdateExportedPublicKeys(ctx, rs.IDs())
	if err != nil {
		out.Errorf(ctx, "Failed to export missing public keys for %q: %s", name, err)

		return err
	}

	// only run second push if we did export any keys.
	if !exported {
		return nil
	}

	if err := sub.Storage().Push(ctx, "", ""); err != nil {
		out.Errorf(ctx, "Failed to push %q to its remote: %s", name, err)

		return err
	}

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
