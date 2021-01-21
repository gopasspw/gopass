package action

import (
	"github.com/gopasspw/gopass/internal/audit"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/tree"
	"github.com/gopasspw/gopass/pkg/ctxutil"

	"github.com/urfave/cli/v2"
)

// Audit validates passwords against common flaws
func (s *Action) Audit(c *cli.Context) error {
	filter := c.Args().First()

	ctx := ctxutil.WithGlobalFlags(c)

	out.Print(ctx, "Auditing passwords for common flaws ...")

	t, err := s.Store.Tree(ctx)
	if err != nil {
		return ExitError(ExitList, err, "failed to get store tree: %s", err)
	}
	if filter != "" {
		subtree, err := t.FindFolder(filter)
		if err != nil {
			return ExitError(ExitUnknown, err, "failed to find subtree: %s", err)
		}
		t = subtree
	}
	list := t.List(tree.INF)

	if len(list) < 1 {
		out.Print(ctx, "No secrets found")
		return nil
	}

	return audit.Batch(ctx, list, s.Store)
}
