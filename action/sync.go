package action

import (
	"context"
	"fmt"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/store"
	"github.com/justwatchcom/gopass/utils/notify"
	"github.com/justwatchcom/gopass/utils/out"
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

	numEntries := 0
	if l, err := s.Store.Tree(); err == nil {
		numEntries = len(l.List(0))
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
	if l, err := s.Store.Tree(); err == nil {
		numEntries = len(l.List(0)) - numEntries
	}
	diff := ""
	if numEntries > 0 {
		diff = fmt.Sprintf(" Added %d entries", numEntries)
	} else if numEntries < 0 {
		diff = fmt.Sprintf(" Removed %d entries", -1*numEntries)
	}
	_ = notify.Notify("gopass - sync", fmt.Sprintf("Finished. Synced %d remotes.%s", numMPs, diff))

	return nil
}

func (s *Action) syncMount(ctx context.Context, mp string) error {
	ctxno := out.WithNewline(ctx, false)
	name := mp
	if mp == "" {
		name = "<root>"
	}
	out.Print(ctxno, color.GreenString("[%s] ", name))

	sub, err := s.Store.GetSubStore(mp)
	if err != nil {
		out.Red(ctx, "Failed to get sub store '%s': %s", name, err)
		return fmt.Errorf("failed to get sub stores (%s)", err)
	}

	if sub == nil {
		out.Red(ctx, "Failed to get sub stores '%s: nil'", name)
		return fmt.Errorf("failed to get sub stores (nil)")
	}

	numMP := 0
	if l, err := sub.List(""); err == nil {
		numMP = len(l)
	}

	out.Print(ctxno, "\n   "+color.GreenString("git pull and push ... "))
	if err := sub.GitPush(ctx, "", ""); err != nil {
		if errors.Cause(err) == store.ErrGitNoRemote {
			out.Yellow(ctx, "Skipped (no remote)")
			out.Debug(ctx, "Failed to push '%s' to it's remote: %s", name, err)
			return err
		}

		out.Red(ctx, "Failed to push '%s' to it's remote: %s", name, err)
		return err
	}
	out.Print(ctxno, color.GreenString("OK"))

	if l, err := sub.List(""); err == nil {
		diff := len(l) - numMP
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
		out.Red(ctx, "Failed to import missing public keys for '%s': %s", name, err)
		return err
	}
	out.Print(ctxno, color.GreenString("OK"))

	// export keys
	out.Print(ctxno, "\n   "+color.GreenString("exporting missing keys ... "))
	rs, err := sub.GetRecipients(ctx, "")
	if err != nil {
		out.Red(ctx, "Failed to load recipients for '%s': %s", name, err)
		return err
	}
	exported, err := sub.ExportMissingPublicKeys(ctx, rs)
	if err != nil {
		out.Red(ctx, "Failed to export missing public keys for '%s': %s", name, err)
		return err
	}

	// only run second push if we did export any keys
	if exported {
		if err := sub.GitPush(ctx, "", ""); err != nil {
			out.Red(ctx, "Failed to push '%s' to it's remote: %s", name, err)
			return err
		}
		out.Print(ctxno, color.GreenString("OK"))
	} else {
		out.Print(ctxno, color.GreenString("nothing to do"))
	}
	out.Print(ctx, "\n   "+color.GreenString("done"))
	return nil
}
