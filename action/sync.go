package action

import (
	"context"
	"fmt"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/store"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

// Sync all stores with their remotes
func (s *Action) Sync(ctx context.Context, c *cli.Context) error {
	out.Green(ctx, "Sync starting ...")

	numEntries := 0
	if l, err := s.Store.Tree(); err == nil {
		numEntries = len(l.List(0))
	}
	numMPs := 0

	mps := s.Store.MountPoints()
	mps = append([]string{""}, mps...)

	for _, mp := range mps {
		numMPs++

		name := mp
		if mp == "" {
			name = "<root>"
		}
		fmt.Print(color.GreenString("[%s] ", name))

		sub, err := s.Store.GetSubStore(mp)
		if err != nil {
			out.Red(ctx, "Failed to get sub store '%s': %s", name, err)
			continue
		}
		if sub == nil {
			out.Red(ctx, "Failed to get sub stores '%s: nil'", name)
			continue
		}

		numMP := 0
		if l, err := sub.List(""); err == nil {
			numMP = len(l)
		}

		fmt.Print("\n   " + color.GreenString("git pull and push ... "))
		if err := sub.GitPush(ctx, "", ""); err != nil {
			if errors.Cause(err) == store.ErrGitNoRemote {
				out.Yellow(ctx, "Skipped (no remote)")
				out.Debug(ctx, "Failed to push '%s' to it's remote: %s", name, err)
				continue
			}
			out.Red(ctx, "Failed to push '%s' to it's remote: %s", name, err)
			continue
		}
		fmt.Print(color.GreenString("OK"))

		if l, err := sub.List(""); err == nil {
			diff := numMP - len(l)
			if diff > 0 {
				fmt.Print(color.GreenString(" (Added %d entries)", diff))
			} else if diff < 0 {
				fmt.Print(color.GreenString(" (Removed %d entries)", diff))
			} else {
				fmt.Print(color.GreenString(" (no changes)"))
			}
		}

		// import keys
		fmt.Print("\n   " + color.GreenString("importing missing keys ... "))
		if err := sub.ImportMissingPublicKeys(ctx); err != nil {
			out.Red(ctx, "Failed to import missing public keys for '%s': %s", name, err)
			continue
		}
		fmt.Print(color.GreenString("OK"))

		// export keys
		fmt.Print("\n   " + color.GreenString("exporting missing keys ... "))
		rs, err := sub.GetRecipients(ctx, "")
		if err != nil {
			out.Red(ctx, "Failed to load recipients for '%s': %s", name, err)
			continue
		}
		exported, err := sub.ExportMissingPublicKeys(ctx, rs)
		if err != nil {
			out.Red(ctx, "Failed to export missing public keys for '%s': %s", name, err)
			continue
		}

		// only run second push if we did export any keys
		if exported {
			if err := sub.GitPush(ctx, "", ""); err != nil {
				out.Red(ctx, "Failed to push '%s' to it's remote: %s", name, err)
				continue
			}
			fmt.Print(color.GreenString("OK"))
		} else {
			fmt.Print(color.GreenString("nothing to do"))
		}
		fmt.Println("\n   " + color.GreenString("done"))
	}
	out.Green(ctx, "All done")

	if l, err := s.Store.Tree(); err == nil {
		numEntries = numEntries - len(l.List(0))
	}
	diff := ""
	if numEntries > 0 {
		diff = fmt.Sprintf(" Added %d entries", numEntries)
	} else if numEntries < 0 {
		diff = fmt.Sprintf(" Removed %d entries", numEntries)
	}
	_ = s.desktopNotify(ctx, "gopass - sync", fmt.Sprintf("Finished. Synced %d remotes.%s", numMPs, diff))

	return nil
}
