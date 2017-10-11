package action

import (
	"context"
	"fmt"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/urfave/cli"
)

// Sync all stores with their remotes
func (s *Action) Sync(ctx context.Context, c *cli.Context) error {
	out.Green(ctx, "Sync starting ...")

	mps := s.Store.MountPoints()
	mps = append([]string{""}, mps...)

	for _, mp := range mps {
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

		fmt.Print("\n   " + color.GreenString("git pull and push ... "))
		if err := sub.GitPush(ctx, "", ""); err != nil {
			out.Red(ctx, "Failed to push '%s' to it's remote: %s", name, err)
			continue
		}
		fmt.Print(color.GreenString("OK"))

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
	return nil
}
