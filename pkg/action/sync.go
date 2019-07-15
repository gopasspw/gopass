package action

import (
	"context"
	"fmt"

	"github.com/gopasspw/gopass/pkg/backend/rcs/noop"
	"github.com/gopasspw/gopass/pkg/notify"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/pkg/store"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

// Sync all stores with their remotes
func (s *Action) Sync(ctx context.Context, c *cli.Context) error {
	store := c.String("store")
	return s.sync(ctx, c, store)
}
func (s *Action) sync(ctx context.Context, c *cli.Context, store string) error {
	out.Green(ctx, "Sync starting ...")

	numMPs := 0

	mps := s.Store.MountPoints()
	mps = append([]string{""}, mps...)
	changes := make(map[string]int)
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
		if syncResult, _ := s.syncMount(ctx, mp); syncResult != nil {
			changes[syncResult.name] = syncResult.changed
		}
	}
	out.Green(ctx, "All done")

	// calculate number of changes entries
	totalAdded := 0
	totalRemoved := 0
<<<<<<< HEAD
	if len(changes) > 0 {
		out.Print(ctx, "Sync summary:\n")
	}
=======
	out.Print(ctx, "Sync summary:\n")
>>>>>>> issue#1134
	for k := range changes {
		if changes[k] > 0 {
			out.Print(ctx, fmt.Sprintf("Mount %s: Added %d entries", k, changes[k]))
			totalAdded += changes[k]
		} else if changes[k] < 0 {
			out.Print(ctx, fmt.Sprintf("Mount %s: Removed %d entries", k, -1*changes[k]))
			totalRemoved += -1 * changes[k]
		} else {
<<<<<<< HEAD
			out.Print(ctx, fmt.Sprintf("Mount %s: no changes", k))
=======
			out.Print(ctx, fmt.Sprintf("Mount %s: didn't change", k))
>>>>>>> issue#1134
		}
	}
	diff := ""
	if totalAdded > 0 {
		diff += fmt.Sprintf(" Added %d entries", totalAdded)
	}
	if totalRemoved > 0 {
<<<<<<< HEAD
		diff += fmt.Sprintf("\nRemoved %d entries", totalRemoved)
=======
		diff += fmt.Sprintf(" Removed %d entries", totalRemoved)
>>>>>>> issue#1134
	}
	_ = notify.Notify(ctx, "gopass - sync", fmt.Sprintf("Finished. Synced %d remotes.%s", numMPs, diff))

	return nil
}

// syncMount syncs a single mount
func (s *Action) syncMount(ctx context.Context, mp string) (*syncResult, error) {
	ctxno := out.WithNewline(ctx, false)
	name := mp
	if mp == "" {
		name = "<root>"
	}
	out.Print(ctxno, color.GreenString("[%s] ", name))

	sub, err := s.Store.GetSubStore(mp)
	if err != nil {
		out.Error(ctx, "Failed to get sub store '%s': %s", name, err)
		return nil, fmt.Errorf("failed to get sub stores (%s)", err)
	}

	if sub == nil {
		out.Error(ctx, "Failed to get sub stores '%s: nil'", name)
		return nil, fmt.Errorf("failed to get sub stores (nil)")
	}

	numMP := 0
	if l, err := sub.List(ctx, ""); err == nil {
		numMP = len(l)
	}

	if sub.RCS().Name() == noop.New().Name() {
		out.Error(ctxno, "\n   WARNING: Mount uses RCS backend 'noop'. Not syncing!\n")
		return nil, nil
	}

	out.Print(ctxno, "\n   "+color.GreenString("git pull and push ... "))
	if err := sub.RCS().Push(ctx, "", ""); err != nil {
		if errors.Cause(err) == store.ErrGitNoRemote {
			out.Yellow(ctx, "Skipped (no remote)")
			out.Debug(ctx, "Failed to push '%s' to its remote: %s", name, err)
			return nil, err
		}

		out.Error(ctx, "Failed to push '%s' to its remote: %s", name, err)
		return nil, err
	}
	out.Print(ctxno, color.GreenString("OK"))
<<<<<<< HEAD
	syncResult := &syncResult{name, 0}
=======
	syncResult := &syncResult{name: name}
>>>>>>> issue#1134
	if l, err := sub.List(ctx, ""); err == nil {
		diff := len(l) - numMP
		syncResult.changed = diff
		if diff > 0 {
			out.Print(ctxno, color.GreenString(" (Added %d entries)", diff))
		} else if diff < 0 {
			out.Print(ctxno, color.GreenString(" (Removed %d entries)", -1*diff))
		} else {
			out.Print(ctxno, color.GreenString(" (no changes)"))
		}
	}

	// import keys
	out.Print(ctxno, "\n   "+color.GreenString("importing missing keys ... "))
	if err := sub.ImportMissingPublicKeys(ctx); err != nil {
		out.Error(ctx, "Failed to import missing public keys for '%s': %s", name, err)
		return syncResult, err
	}
	out.Print(ctxno, color.GreenString("OK"))

	// export keys
	out.Print(ctxno, "\n   "+color.GreenString("exporting missing keys ... "))
	rs, err := sub.GetRecipients(ctx, "")
	if err != nil {
		out.Error(ctx, "Failed to load recipients for '%s': %s", name, err)
		return syncResult, err
	}
	exported, err := sub.ExportMissingPublicKeys(ctx, rs)
	if err != nil {
		out.Error(ctx, "Failed to export missing public keys for '%s': %s", name, err)
		return syncResult, err
	}

	// only run second push if we did export any keys
	if exported {
		if err := sub.RCS().Push(ctx, "", ""); err != nil {
			out.Error(ctx, "Failed to push '%s' to its remote: %s", name, err)
			return syncResult, err
		}
		out.Print(ctxno, color.GreenString("OK"))
	} else {
		out.Print(ctxno, color.GreenString("nothing to do"))
	}
	out.Print(ctx, "\n   "+color.GreenString("done"))
	return syncResult, nil
}

type syncResult struct {
	name    string
	changed int
}
