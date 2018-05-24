package action

import (
	"context"

	"github.com/justwatchcom/gopass/pkg/audit"
	"github.com/justwatchcom/gopass/pkg/out"

	"github.com/urfave/cli"
)

// Audit validates passwords against common flaws
func (s *Action) Audit(ctx context.Context, c *cli.Context) error {
	filter := c.Args().First()

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
