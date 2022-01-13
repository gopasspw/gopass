package action

import (
	"os"
	"os/exec"
	"strings"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/urfave/cli/v2"
)

// Git passes the git command to the underlying backend.
func (s *Action) Git(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	store := c.String("store")

	sub, err := s.Store.GetSubStore(store)
	if err != nil || sub == nil {
		return ExitError(ExitGit, err, "failed to get sub store %s: %s", store, err)
	}

	args := c.Args().Slice()
	out.Noticef(ctx, "Running 'git %s' in %s...", strings.Join(args, " "), sub.Path())
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = sub.Path()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
