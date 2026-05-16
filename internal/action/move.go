package action

import (
	"context"
	"fmt"

	"github.com/gopasspw/gopass/internal/action/exit"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/termio"
	"github.com/urfave/cli/v3"
)

// Move the content from one secret to another.
func (s *secretHandler) Move(ctx context.Context, cmd *cli.Command) error {
	ctx = ctxutil.WithGlobalFlags(ctx, cmd)

	if cmd.Args().Len() != 2 {
		return exit.Error(exit.Usage, nil, "Usage: %s mv old-path new-path", s.Name)
	}

	from := cmd.Args().Get(0)
	to := cmd.Args().Get(1)

	if !cmd.Bool("force") {
		if s.Store.Exists(ctx, to) && !termio.AskForConfirmation(ctx, fmt.Sprintf("%s already exists. Overwrite it?", to)) {
			return exit.Error(exit.Aborted, nil, "not overwriting your current secret")
		}
	}

	// Check for custom commit message
	commitMsg := fmt.Sprintf("Moved %s to %s", from, to)
	if cmd.IsSet("commit-message") {
		commitMsg = cmd.String("commit-message")
	}
	if cmd.Bool("interactive-commit") {
		commitMsg = ""
	}
	ctx = ctxutil.WithCommitMessage(ctx, commitMsg)

	if err := s.Store.Move(ctx, from, to); err != nil {
		return exit.Error(exit.Unknown, err, "%s", err)
	}

	return nil
}
