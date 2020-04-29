package action

import (
	"github.com/gopasspw/gopass/pkg/audit"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/out"

	"github.com/urfave/cli/v2"
)

// Audit validates passwords against common flaws
func (s *Action) Audit(c *cli.Context) error {
	filter := c.Args().First()

	ctx := ctxutil.WithGlobalFlags(c)
	ctx = s.Store.WithConfig(ctx, filter)

	out.Print(ctx, "Auditing passwords for common flaws ...")

	t, err := s.Store.Tree(ctx)
	if err != nil {
		return ExitError(ctx, ExitList, err, "failed to get store tree: %s", err)
	}
	if filter != "" {
		subtree, err := t.FindFolder(filter)
		if err != nil {
			return ExitError(ctx, ExitUnknown, err, "failed to find subtree: %s", err)
		}
		t = subtree
	}
	list := t.List(0)

	if len(list) < 1 {
		out.Yellow(ctx, "No secrets found")
		return nil
	}

	return audit.Batch(ctx, list, s.Store)
}
